package postgres

import (
	"fmt"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/jmoiron/sqlx"
)

// Connect creates a new PostgreSQL database connection using the provided configuration.
//
// The function establishes a connection to the database, applies connection pool
// settings, and returns a ready-to-use sqlx.DB instance.
//
// It returns an error if the database connection cannot be established.
func Connect(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", getDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with db: %w", err)
	}

	applySettingsDB(db, cfg)
	return db, nil
}

func applySettingsDB(db *sqlx.DB, cfg *config.Config) {
	db.SetMaxOpenConns(cfg.Persistence.Postgres.Conns.MaxOpens)
	db.SetMaxIdleConns(cfg.Persistence.Postgres.Conns.MaxIdles)
	db.SetConnMaxIdleTime(cfg.Persistence.Postgres.Conns.MaxIdleTime)
	db.SetConnMaxLifetime(cfg.Persistence.Postgres.Conns.MaxLifetime)
}

func getDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d sslmode=%s user=%s password=%s dbname=%s",
		cfg.Persistence.Postgres.Host,
		cfg.Persistence.Postgres.Port,
		cfg.Persistence.Postgres.SSLMode,
		cfg.Persistence.Postgres.Auth.User,
		cfg.Persistence.Postgres.Auth.Password,
		cfg.Persistence.Postgres.Auth.DBName,
	)
}
