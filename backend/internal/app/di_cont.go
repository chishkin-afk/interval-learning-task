package app

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/http/server"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/http/server/handlers"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/persistence/postgres"
	logger "github.com/chishkin-afk/intask/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// DI is a dependency injection container
//
// It is designed for the safe invocation
// of functions with once-only initialization.
// Calling its methods implies that .env is loaded.
type DI struct {
	cfg *config.Config
	log *slog.Logger

	db *sqlx.DB

	handler *gin.Engine
	server  *server.Server
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

func (di *DI) DB() *sqlx.DB {
	if di.db == nil {
		db, err := postgres.Connect(di.Config())
		if err != nil {
			slog.Error("failed to connect db",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		di.db = db
	}

	return di.db
}

func (di *DI) Handler() *gin.Engine {
	if di.handler == nil {
		di.handler = handlers.New(
			di.Config(),
		)
	}

	return di.handler
}

func (di *DI) Server() *server.Server {
	if di.server == nil {
		di.server = server.New(
			di.Config(),
			di.Handler(),
		)
	}

	return di.server
}
