package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
	"gopkg.in/telebot.v3"
)

// Client is a high-level wrapper around telebot.Bot that adds structured
// logging, safe message delivery, and lifecycle management.
//
// A Client is safe for concurrent use after Start has been called, except
// for Handle and Use, which must be invoked before Start.
type Client struct {
	log *slog.Logger
	bot *telebot.Bot
}

// New creates a new Client using the provided configuration and logger.
// It establishes a connection to the Telegram Bot API via long polling.
//
// New returns an error if the bot token is invalid or the Telegram API
// is unreachable at construction time. The returned Client is not yet
// polling for updates; call Start to begin receiving events.
//
// cfg.Telegram.Token must be a valid bot token obtained from BotFather.
// cfg.Telegram.Poll.Timeout controls the long-polling timeout passed to
// the underlying telebot.LongPoller.
func New(cfg *config.Config, log *slog.Logger) (*Client, error) {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.Telegram.Token,
		Poller: &telebot.LongPoller{Timeout: cfg.Telegram.Poll.Timeout},
	})
	if err != nil {
		return nil, fmt.Errorf("can't open conn with tg: %w", err)
	}

	return &Client{
		log: log,
		bot: bot,
	}, nil
}

// Handle registers a handler for the given endpoint (command, callback,
// or message type). It is a thin, logged wrapper around telebot.Bot.Handle.
//
// Handle must be called before Start; registering handlers after the bot
// has started leads to undefined behavior.
//
// The endpoint argument accepts any value supported by telebot (e.g. "/start",
// telebot.OnText, a Callback unique string).
func (c *Client) Handle(endpoint any, handler telebot.HandlerFunc) {
	c.log.Debug("new handle to endpoint",
		slog.Any("endpoint", endpoint),
	)

	c.bot.Handle(endpoint, handler)
}

// SendString sends a plain text message to the chat identified by chatID.
//
// SendString silently ignores errors caused by an inactive chat
// (the user blocked the bot or the chat no longer exists) and logs them
// at Warn level, returning nil. This allows callers to treat such chats
// as candidates for cleanup without aborting batch deliveries.
//
// Any other transport or API error is wrapped and returned to the caller.
func (c *Client) SendString(ctx context.Context, msg string, chatID int64) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if _, err := c.bot.Send(telebot.ChatID(chatID), msg); err != nil {
		if errors.Is(err, telebot.ErrBlockedByUser) ||
			errors.Is(err, telebot.ErrChatNotFound) {
			c.log.Warn("inactive chat, consider removing",
				slog.Int64("chat_id", chatID),
				slog.String("error", err.Error()),
			)
			return nil
		}
		return err
	}

	return nil
}

// Use registers one or more middlewares that will be applied to every
// incoming update in the order they are provided.
//
// Use must be called before Start; adding middlewares after the bot has
// started has no effect on already-dispatched updates.
func (c *Client) Use(mws ...telebot.MiddlewareFunc) {
	c.log.Debug("setting middlewares",
		slog.Int("count", len(mws)),
	)

	c.bot.Use(mws...)
}

// Start begins long-polling the Telegram Bot API for updates. It blocks
// until Stop is called or a fatal error occurs inside the underlying
// telebot.Bot.
//
// Start is typically invoked in a separate goroutine:
//
//	go client.Start()
//
// All handlers and middlewares must be registered before calling Start.
func (c *Client) Start() {
	c.log.Info("bot has been started...")
	c.bot.Start()
}

// Stop gracefully shuts down the long-polling loop initiated by Start.
// It is safe to call Stop multiple times; subsequent calls are no-ops
// at the telebot layer.
//
// Stop should be paired with Start via defer to ensure clean shutdown:
//
//	go client.Start()
//	defer client.Stop()
func (c *Client) Stop() {
	c.log.Info("stopping bot")

	done := make(chan struct{})

	go func() {
		c.bot.Stop()
		close(done)
	}()

	select {
	case <-done:
		c.log.Info("bot has been stopped")
	case <-time.After(15 * time.Second):
		c.log.Warn("bot stop timed out, forcing exit")
		return
	}
}
