package userpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/chishkin-afk/intask/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	allColumns = []string{
		"id",
		"tg_chat_id",
		"telegram_enabled",
		"email",
		"password_hash",
		"created_at",
		"updated_at",
	}
)

// DB represents the minimal database functionality required by the user
// repository to execute SQL statements and retrieve rows.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type scanner interface {
	Scan(dest ...any) error
}

// WithTx extends DB with transaction support.
type TxDB interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type userRepository struct {
	log        *slog.Logger
	db         TxDB
	sqlBuilder squirrel.StatementBuilderType
}

// New creates a new PostgreSQL-backed user repository.
func New(log *slog.Logger, db TxDB) *userRepository {
	return &userRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save persists the provided user in the database.
//
// Returns an error if the user cannot be stored or a user with the same
// identifier already exists.
func (ur *userRepository) Save(ctx context.Context, user *user.User) error {
	ur.log.Debug("saving user into db",
		slog.String("user_id", user.ID().String()),
	)

	if err := ur.save(ctx, ur.db, user); err != nil {
		return fmt.Errorf("failed to save user into db: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) save(ctx context.Context, db DB, user *user.User) error {
	query, args, err := ur.buildInsertQuery(userToRecord(user))
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) buildInsertQuery(record *userRecord) (string, []any, error) {
	return ur.sqlBuilder.Insert("users").Columns(
		allColumns...,
	).Values(
		record.ID,
		record.TgChatID,
		record.TgEnabled,
		record.Email,
		record.PasswordHash,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

// GetByID retrieves a user by its unique identifier.
//
// Returns user.ErrNotFound if no user with the given identifier exists.
// Any other error indicates a failure while accessing the database.
func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ur.log.Debug("getting user by id",
		slog.String("user_id", id.String()),
	)

	user, err := ur.getByID(ctx, ur.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from db: %w",
			handleError(err),
		)
	}

	return user, nil
}

func (ur *userRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*user.User, error) {
	query, args, err := ur.buildSelectByIDQuery(id)
	if err != nil {
		return nil, err
	}

	return scanUser(db.QueryRowContext(ctx, query, args...))
}

func (ur *userRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return ur.sqlBuilder.Select(allColumns...).From("users").
		Where("id = ?", id).ToSql()
}

// GetByEmail retrieves a user by its unique email.
//
// Returns user.ErrNotFound if no user with the given email exists.
// Any other error indicates a failure while accessing the database.
func (ur *userRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	ur.log.Debug("getting user by email",
		slog.String("email", email.String()),
	)

	user, err := ur.getByEmail(ctx, ur.db, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w",
			handleError(err),
		)
	}

	return user, nil
}

func (ur *userRepository) getByEmail(ctx context.Context, db DB, email user.Email) (*user.User, error) {
	query, args, err := ur.buildSelectByEmailQuery(email)
	if err != nil {
		return nil, err
	}

	return scanUser(db.QueryRowContext(ctx, query, args...))
}

func (ur *userRepository) buildSelectByEmailQuery(email user.Email) (string, []any, error) {
	return ur.sqlBuilder.Select(allColumns...).From("users").
		Where("email = ?", email).ToSql()
}

// Update applies the provided update function to the user identified by id
// and persists the changes atomically.
//
// Returns user.ErrNotFound if no user with the given identifier exists.
func (ur *userRepository) Update(ctx context.Context, id uuid.UUID, updFunc user.UpdateFunc) (*user.User, error) {
	ur.log.Debug("starting tx for update")
	tx, rollback, err := ur.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer rollback()

	ur.log.Debug("getting user for update",
		slog.String("user_id", id.String()),
	)

	user, err := ur.getForUpdate(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w",
			handleError(err),
		)
	}

	ur.log.Debug("calling update func...")
	if err := updFunc(ctx, user); err != nil {
		return nil, err
	}

	if err := ur.update(ctx, tx, user); err != nil {
		return nil, fmt.Errorf("failed to update user in db: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return user, nil
}

func (ur *userRepository) update(ctx context.Context, db DB, updUser *user.User) error {
	query, args, err := ur.buildUpdateQuery(userToRecord(updUser))
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (ur *userRepository) buildUpdateQuery(record *userRecord) (string, []any, error) {
	return ur.sqlBuilder.Update("users").SetMap(map[string]any{
		"tg_chat_id":       record.TgChatID,
		"telegram_enabled": record.TgEnabled,
		"updated_at":       record.UpdatedAt,
	}).Where("id = ?", record.ID).ToSql()
}

func (ur *userRepository) getForUpdate(ctx context.Context, db DB, id uuid.UUID) (*user.User, error) {
	query, args, err := ur.buildSelectForUpdate(id)
	if err != nil {
		return nil, err
	}

	return scanUser(db.QueryRowContext(ctx, query, args...))
}

func (ur *userRepository) buildSelectForUpdate(id uuid.UUID) (string, []any, error) {
	return ur.sqlBuilder.Select(allColumns...).From("users").
		Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

// Delete removes the user identified by id from the database.
//
// Returns user.ErrNotFound if the user does not exist.
// Any other error indicates a failure while accessing the database.
func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ur.log.Debug("deleting user from db",
		slog.String("user_id", id.String()),
	)

	if err := ur.delete(ctx, ur.db, id); err != nil {
		return fmt.Errorf("failed to delete user from db: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := ur.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return user.ErrNotFound
	}

	return nil
}

func (ur *userRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return ur.sqlBuilder.Delete("users").Where("id = ?", id).ToSql()
}

func (ur *userRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := ur.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			ur.log.Error("failed to rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func scanUser(row *sql.Row) (*user.User, error) {
	record, err := scan(row)
	if err != nil {
		return nil, err
	}

	return recordToUser(record), nil
}

func scan(scnr scanner) (*userRecord, error) {
	var record userRecord
	if err := scnr.Scan(
		&record.ID,
		&record.TgChatID,
		&record.TgEnabled,
		&record.Email,
		&record.PasswordHash,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &record, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user.ErrNotFound
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		switch pqErr.Code {
		case "23505":
			return user.ErrAlreadyExists
		}
	}

	return err
}
