package services

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/chishkin/intask/notifier/internal/application/dtos/requests"
	"github.com/chishkin/intask/notifier/internal/domain/user"
	"github.com/chishkin/intask/notifier/pkg/errs"
	"github.com/google/uuid"
)

const (
	MinLenMessage = 3
	MaxLenMessage = 128
)

type TgBot interface {
	SendString(ctx context.Context, msg string, chatID int64) error
	OnStart(h func(ctx context.Context, payload string, chatID int64))
}

type BackendClient interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	BindTg(ctx context.Context, code, chatID int64) error
}

type NotifierService struct {
	log           *slog.Logger
	tgbot         TgBot
	backendClient BackendClient
}

func New(
	log *slog.Logger,
	tgbot TgBot,
	backendClient BackendClient,
) *NotifierService {
	return &NotifierService{
		log:           log,
		tgbot:         tgbot,
		backendClient: backendClient,
	}
}

// RegisterTgHandlers wires service methods to bot commands.
// Must be called before tgbot.Client.Start.
func (ns *NotifierService) RegisterTgHandlers() {
	ns.tgbot.OnStart(ns.onStart)
}

func (ns *NotifierService) onStart(ctx context.Context, payload string, chatID int64) {
	ns.log.Debug("new user on start")

	code, err := strconv.ParseInt(payload, 0, 64)
	if err != nil {
		ns.log.Error("can't parse code from bot",
			slog.String("error", err.Error()),
			slog.String("payload", payload),
		)

		ns.sendMsg(ctx, chatID, "try to follow link on site again!")
		return
	}

	if err := ns.backendClient.BindTg(ctx, int64(code), chatID); err != nil {
		ns.log.Error("can't bind tg on backend",
			slog.String("error", err.Error()),
		)

		ns.sendMsg(ctx, chatID, "ooops, something wrong. please, try again later...")
	}
}

// SendMsg validates req, resolves the recipient, and sends msg via Telegram.
//
// It returns a KindError; use errors.As to inspect the Kind.
func (ns *NotifierService) SendMsg(ctx context.Context, req *requests.SendMsg) error {
	msg := strings.TrimSpace(req.Msg)
	if err := validateMsg(msg); err != nil {
		return err
	}

	u, err := ns.getUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	if !u.TgEnabled() {
		return errs.NewKindError(errs.KindPermissionDenied, errs.ErrTgDisabled)
	}

	return ns.sendMsg(ctx, u.TgChatID(), msg)
}

func (ns *NotifierService) sendMsg(ctx context.Context, tgChatID int64, msg string) error {
	if err := ns.tgbot.SendString(ctx, msg, tgChatID); err != nil {
		ns.log.Error("failed to send msg",
			slog.String("error", err.Error()),
			slog.Int64("tg_chat_id", tgChatID),
		)

		return handleError(err)
	}

	return nil
}

func (ns *NotifierService) getUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	user, err := ns.backendClient.GetUserByID(ctx, id)
	if err != nil {
		ns.log.Error("can't get user from backend",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)

		return nil, handleError(err)
	}

	return user, nil
}

func validateMsg(msg string) error {
	n := len([]rune(msg))
	if n < MinLenMessage || n > MaxLenMessage {
		return errs.NewKindError(errs.KindRequest, errs.ErrInvalidMessage)
	}

	return nil
}

func handleError(err error) error {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return errs.NewKindError(errs.KindTimeout, err)
	}

	return errs.NewKindError(errs.KindInternal, errs.ErrInternalServer)
}
