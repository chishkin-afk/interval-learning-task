package app

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	logger "github.com/chishkin-afk/intask/backend/pkg/log"
)

// DI is a dependency injection container
//
// It is designed for the safe invocation
// of functions with once-only initialization.
// Calling its methods implies that .env is loaded.
type DI struct {
	cfg *config.Config
	log *slog.Logger
}

func (di *DI) Config() *config.Config {
	if di.cfg == nil {
		cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
		if err != nil {
			slog.Error("failed to load config",
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
		di.log = slog.New(logger.NewHandler(
			di.Config().App.Env,
		))
	}

	return di.log
}
