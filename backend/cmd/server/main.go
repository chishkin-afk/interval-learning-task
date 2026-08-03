package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/chishkin-afk/intask/backend/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()),
		)
	}

	var di app.DI

	go func() {
		errs := di.WorkerPool().Errors()
		for err := range errs {
			di.Log().Error("err in worker pool",
				slog.String("error", err.Error()),
			)
		}

		di.Log().Info("goroutine with errs wp was closed",
			slog.Int64("dropped_errors", di.WorkerPool().Dropped()),
		)
	}()

	go func() {
		di.Log().Info("server is running...",
			slog.String("addr", di.Config().Server.HTTP.Addr),
		)

		if err := di.Server().Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			di.Log().Error("failed to start server",
				slog.String("error", err.Error()),
			)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.CloseAll(shutdownCtx); err != nil {
		di.Log().Error("failed to close all deps",
			slog.String("error", err.Error()),
		)
	}

	di.Log().Info("all the deps were closed")
}
