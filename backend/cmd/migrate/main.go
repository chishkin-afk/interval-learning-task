package main

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/intask/backend/internal/app"
	"github.com/chishkin-afk/intask/backend/internal/infrastructure/persistence/postgres"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()),
		)
	}

	var di app.DI
	di.Log().Info("starts to migrate up...",
		slog.String("migrations_path", di.Config().Persistence.MigrationsPath),
	)

	if err := postgres.Migrate(di.Config()); err != nil {
		di.Log().Error("failed to migrate db",
			slog.String("error", err.Error()),
			slog.String("migrations_path", di.Config().Persistence.MigrationsPath),
		)
		os.Exit(1)
	}

	di.Log().Info("all the migrations has been applied.")
}
