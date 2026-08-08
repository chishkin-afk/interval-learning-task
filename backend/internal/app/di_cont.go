package app

import (
	"context"
	"log/slog"
	"os"

	redisconnect "github.com/chishkin-afk/intask/backend/internal/infrastructure/cache/redis"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/http/server"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/http/server/handlers"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/http/server/middlewares"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/jwt"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/persistence/postgres"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/workerpool"
	authservice "github.com/chishkin-afk/intask/backend/internal/modules/auth/application/services"
	userredis "github.com/chishkin-afk/intask/backend/internal/modules/auth/infrastructure/cache/redis/user"
	userpg "github.com/chishkin-afk/intask/backend/internal/modules/auth/infrastructure/persistence/postgres/user"
	taskservice "github.com/chishkin-afk/intask/backend/internal/modules/task/application/services"
	taskpg "github.com/chishkin-afk/intask/backend/internal/modules/task/infrastructure/persistence/postgres/task"
	logger "github.com/chishkin-afk/intask/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// DI is a dependency injection container
//
// It is designed for the safe invocation
// of functions with once-only initialization.
// Calling its methods implies that .env is loaded.
type DI struct {
	cfg *config.Config
	log *slog.Logger

	db    *sqlx.DB
	cache *redis.Client

	handler *gin.Engine
	server  *server.Server

	jwtMngr *jwt.JWTManager

	authService *authservice.AuthService
	taskService *taskservice.TaskService

	workerpool *workerpool.WorkerPool
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

		Add(func(ctx context.Context) error {
			errCh := make(chan error, 1)
			go func() {
				defer close(errCh)
				if err := db.Close(); err != nil {
					errCh <- err
				}
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err, ok := <-errCh:
				if !ok {
					return err
				}

				return nil
			}
		})

		di.db = db
	}

	return di.db
}

func (di *DI) Handler() *gin.Engine {
	if di.handler == nil {
		di.handler = handlers.New(
			di.Config(),
			di.AuthService(),
			di.TaskService(),
			middlewares.NewAuthMiddleware(
				di.JWT(),
				map[string]bool{
					"/api/v1/auth/user":      true,
					"/api/v1/auth/code":      true,
					"/api/v1/tasks/task/:id": true,
					"/api/v1/tasks/task":     true,
					"/api/v1/tasks/":         true,
				},
			),
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

		Add(func(ctx context.Context) error {
			return di.server.Shutdown(ctx)
		})
	}

	return di.server
}

func (di *DI) AuthService() *authservice.AuthService {
	if di.authService == nil {
		di.authService = authservice.New(
			di.Config(),
			di.Log(),
			userpg.New(di.Log(), di.DB()),
			userredis.New(
				di.Config(),
				di.Log(),
				di.Cache(),
			),
			di.JWT(),
		)
	}

	return di.authService
}

func (di *DI) TaskService() *taskservice.TaskService {
	if di.taskService == nil {
		di.taskService = taskservice.New(
			di.Config(),
			di.Log(),
			taskpg.New(
				di.Log(),
				di.DB(),
			),
			di.WorkerPool(),
			nil,
		)
	}

	return di.taskService
}

func (di *DI) JWT() *jwt.JWTManager {
	if di.jwtMngr == nil {
		jwtMngr, err := jwt.New(di.Config())
		if err != nil {
			slog.Error("failed to get jwt manager",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		di.jwtMngr = jwtMngr
	}

	return di.jwtMngr
}

func (di *DI) WorkerPool() *workerpool.WorkerPool {
	if di.workerpool == nil {
		di.workerpool = workerpool.New(
			di.Config(),
		)
	}

	return di.workerpool
}

func (di *DI) Cache() *redis.Client {
	if di.cache == nil {
		cache, err := redisconnect.Connect(di.Config())
		if err != nil {
			slog.Error("can't connect cache",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}

		di.cache = cache

		Add(func(ctx context.Context) error {
			errCh := make(chan error, 1)
			go func() {
				defer close(errCh)
				errCh <- di.cache.Close()
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errCh:
				return err
			}
		})
	}

	return di.cache
}
