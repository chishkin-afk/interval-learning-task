package postgres

import (
	"errors"
	"fmt"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate applies migration into db
//
// This function connects to the database temporarily
// and applies migrations specified in the configuration
// otherwise, it returns an error.
func Migrate(cfg *config.Config) error {
	migration, err := migrate.New("file://"+cfg.Persistence.MigrationsPath, getURL(cfg))
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}
	defer migration.Close()

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to up migration: %w", err)
	}

	return nil
}

func getURL(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Persistence.Postgres.Auth.User,
		cfg.Persistence.Postgres.Auth.Password,
		cfg.Persistence.Postgres.Host,
		cfg.Persistence.Postgres.Port,
		cfg.Persistence.Postgres.Auth.DBName,
		cfg.Persistence.Postgres.SSLMode,
	)
}
