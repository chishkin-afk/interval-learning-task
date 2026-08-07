package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/chishkin/intask/notifier/internal/application/services"
	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
	"github.com/chishkin/intask/notifier/internal/infrastructure/http/backend"
	"github.com/chishkin/intask/notifier/internal/infrastructure/http/server"
	"github.com/chishkin/intask/notifier/internal/infrastructure/http/server/handlers"
	"github.com/chishkin/intask/notifier/internal/infrastructure/tgbot"
	logger "github.com/chishkin/intask/notifier/pkg/log"
	"github.com/gin-gonic/gin"
)

type DI struct {
	cfg *config.Config
	log *slog.Logger

	server  *server.Server
	handler *gin.Engine

	tgBot         *tgbot.Client
	backendClient *backend.Client

	notifierService *services.NotifierService
}

func (di *DI) Config() *config.Config {
	if di.cfg == nil {
		cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
		if err != nil {
			slog.Error("can't load config",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		di.cfg = cfg
	}

	return di.cfg
}

func (di *DI) Log() *slog.Logger {
	if di.log == nil {
		di.log = slog.New(logger.NewHandler(di.Config().App.Env))
	}

	return di.log
}

func (di *DI) NotifierService() *services.NotifierService {
	if di.notifierService == nil {
		di.notifierService = services.New(
			di.Log(),
			di.TgBot(),
			di.BackendClient(),
		)
	}

	return di.notifierService
}

func (di *DI) BackendClient() *backend.Client {
	if di.backendClient == nil {
		di.backendClient = backend.New(
			di.Config(),
			di.Log(),
			&http.Client{
				Timeout: di.Config().Backend.Timeout,
			},
		)
	}

	return di.backendClient
}

func (di *DI) TgBot() *tgbot.Client {
	if di.tgBot == nil {
		tgBot, err := tgbot.New(
			di.Config(),
			di.Log(),
		)
		if err != nil {
			slog.Error("can't connect tg bot",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		di.tgBot = tgBot

		Add(func(ctx context.Context) error {
			done := make(chan struct{})
			go func() {
				defer close(done)
				di.tgBot.Stop()
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				return nil
			}
		})
	}

	return di.tgBot
}

func (di *DI) Handler() *gin.Engine {
	if di.handler == nil {
		di.handler = handlers.New(
			di.Config(),
			di.NotifierService(),
		)
	}

	return di.handler
}

func (di *DI) Server() *server.Server {
	if di.server == nil {
		di.server = server.New(
			di.Config(),
			di.Log(),
			di.Handler(),
		)

		Add(func(ctx context.Context) error {
			return di.server.Shutdown(ctx)
		})
	}

	return di.server
}
