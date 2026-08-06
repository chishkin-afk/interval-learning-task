package app

import (
	"log/slog"
	"os"

	"github.com/chishkin/intask/notifier/internal/infrastructure/config"
	logger "github.com/chishkin/intask/notifier/pkg/log"
)

type DI struct {
	cfg *config.Config
	log *slog.Logger
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
