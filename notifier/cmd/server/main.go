package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chishkin/intask/notifier/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("can't load .env",
			slog.String("error", err.Error()),
		)
	}
	var di app.DI
	di.TgBot()

	di.NotifierService().RegisterTgHandlers()

	di.Log().Info("di was loaded...")

	go func() {
		if err := di.Server().Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			di.Log().Error("can't start server",
				slog.String("error", err.Error()),
			)
		}
	}()

	go di.TgBot().Start()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGABRT,
		syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.CloseAll(shutdownCtx); err != nil {
		di.Log().Error("can't close some deps",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	di.Log().Info("all was closed")
}
