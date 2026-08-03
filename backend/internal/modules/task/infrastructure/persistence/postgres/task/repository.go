package taskpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/chishkin-afk/intask/backend/internal/modules/task/domain/task"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var taskColumns = []string{
	"id",
	"user_id",
	"title",
	"leetcode_url",
	"next_notify",
	"notify_count",
	"is_active",
	"created_at",
}

// DB represents the minimal database functionality required by the user
// repository to execute SQL statements and retrieve rows.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type scanner interface {
	Scan(dest ...any) error
}

// WithTx extends DB with transaction support.
type TxDB interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type taskRepository struct {
	log *slog.Logger
	db  TxDB
	sb  squirrel.StatementBuilderType
}

func New(log *slog.Logger, db TxDB) *taskRepository {
	return &taskRepository{
		log: log,
		db:  db,
		sb:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save persists the given task in the database.
//
// Domain-specific errors returned by the database are converted into
// domain errors before being returned.
func (tr *taskRepository) Save(ctx context.Context, task *task.Task) error {
	tr.log.Debug("saving task into db",
		slog.String("task_id", task.ID().String()),
	)

	if err := tr.save(ctx, tr.db, task); err != nil {
		return fmt.Errorf("can't save task into db: %w",
			handleError(err),
		)
	}

	return nil
}

func (tr *taskRepository) save(ctx context.Context, db DB, task *task.Task) error {
	query, args, err := tr.buildInsertQuery(taskToRecord(task))
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a task by its identifier.
//
// Database-specific errors are converted into domain errors before
// being returned.
func (tr *taskRepository) GetByID(ctx context.Context, id uuid.UUID) (*task.Task, error) {
	tr.log.Debug("getting task by id",
		slog.String("task_id", id.String()),
	)

	task, err := tr.getByID(ctx, tr.db, id)
	if err != nil {
		return nil, fmt.Errorf("can't get task from db: %w",
			handleError(err),
		)
	}

	return task, nil
}

func (tr *taskRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*task.Task, error) {
	query, args, err := tr.buildSelectQuery(id)
	if err != nil {
		return nil, err
	}

	return scanTask(db.QueryRowContext(ctx, query, args...))
}

// ListAll returns a paginated list of tasks and the total number of tasks.
//
// The list and the total count are read within a single read-only transaction
// to ensure they are based on the same database snapshot.
func (tr *taskRepository) ListAll(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*task.Task, int64, error) {
	tx, rollback, err := tr.beginReadTx(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("can't start tx: %w", err)
	}
	defer rollback()

	count, err := tr.countAll(ctx, tx)
	if err != nil {
		return nil, 0, fmt.Errorf("can't count tasks: %w",
			handleError(err),
		)
	}

	tr.log.Debug("listing all the tasks",
		slog.Int("page", int(page)),
		slog.Int("limit", int(limit)),
	)

	list, err := tr.listAll(ctx, tx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("can't list tasks: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("can't commit tx: %w", err)
	}

	return list, count, nil
}

func (tr *taskRepository) listAll(ctx context.Context, db DB, userID uuid.UUID, page, limit uint32) ([]*task.Task, error) {
	query, args, err := tr.buildListAllQuery(userID, page, limit)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (tr *taskRepository) countAll(ctx context.Context, db DB) (int64, error) {
	query, args, err := tr.buildCountAllQuery()
	if err != nil {
		return 0, err
	}

	return scanCount(db.QueryRowContext(ctx, query, args...))
}

// ListByNotification returns a paginated list of tasks that are ready
// to be notified and the total number of such tasks.
//
// The list and total count are retrieved within a single read-only transaction
// to ensure they are based on the same database snapshot.
func (tr *taskRepository) ListByNotification(ctx context.Context, page, limit uint32) ([]*task.Task, int64, error) {
	tx, rollback, err := tr.beginReadTx(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("can't start tx: %w", err)
	}
	defer rollback()

	count, err := tr.countByNotification(ctx, tx)
	if err != nil {
		return nil, 0, fmt.Errorf("can't count tasks: %w",
			handleError(err),
		)
	}

	tr.log.Debug("listing all the tasks by notify",
		slog.Int("page", int(page)),
		slog.Int("limit", int(limit)),
	)

	list, err := tr.listByNotification(ctx, tx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("can't list tasks by notify: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("can't commit tx: %w", err)
	}

	return list, count, nil
}

func (tr *taskRepository) listByNotification(ctx context.Context, db DB, page, limit uint32) ([]*task.Task, error) {
	query, args, err := tr.buildListByNotificationQuery(page, limit)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (tr *taskRepository) countByNotification(ctx context.Context, db DB) (int64, error) {
	query, args, err := tr.buildCountByNotificationQuery()
	if err != nil {
		return 0, err
	}

	return scanCount(db.QueryRowContext(ctx, query, args...))
}

// Update retrieves a task by its identifier, applies the provided update
// function, and persists the changes.
//
// The task row is locked for update during the transaction to prevent
// concurrent modifications from overwriting each other.
func (tr *taskRepository) Update(ctx context.Context, id uuid.UUID, updFunc task.UpdateFunc) (*task.Task, error) {
	tx, rollback, err := tr.beginWriteTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't start tx: %w", err)
	}
	defer rollback()

	task, err := tr.getForUpdate(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("can't get task for update: %w",
			handleError(err),
		)
	}

	if err := updFunc(ctx, task); err != nil {
		return nil, err
	}

	tr.log.Debug("updating the task",
		slog.String("task_id", id.String()),
	)

	if err := tr.update(ctx, tx, task); err != nil {
		return nil, fmt.Errorf("can't update task: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("can't commit tx: %w", err)
	}

	return task, nil
}

func (tr *taskRepository) update(ctx context.Context, db DB, updTask *task.Task) error {
	query, args, err := tr.buildUpdateQuery(taskToRecord(updTask))
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return task.ErrNotFound
	}

	return nil
}

func (tr *taskRepository) getForUpdate(ctx context.Context, db DB, id uuid.UUID) (*task.Task, error) {
	query, args, err := tr.buildSelectForUpdateQuery(id)
	if err != nil {
		return nil, err
	}

	return scanTask(db.QueryRowContext(ctx, query, args...))
}

// Delete removes a task with the specified identifier from the database.
//
// Returns task.ErrNotFound if a task with the given identifier does not exist.
func (tr *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tr.log.Debug("deleting task from db",
		slog.String("task_id", id.String()),
	)

	if err := tr.delete(ctx, tr.db, id); err != nil {
		return fmt.Errorf("can't delete task from db: %w",
			handleError(err),
		)
	}

	return nil
}

func (tr *taskRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := tr.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return task.ErrNotFound
	}

	return nil
}

func (tr *taskRepository) beginReadTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := tr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			tr.log.Error("can't rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func (tr *taskRepository) beginWriteTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := tr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			tr.log.Error("can't rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func scanCount(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func scanTasks(rows *sql.Rows) ([]*task.Task, error) {
	var tasks []*task.Task
	for rows.Next() {
		record, err := scan(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, recordToTask(record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func scanTask(row *sql.Row) (*task.Task, error) {
	record, err := scan(row)
	if err != nil {
		return nil, err
	}

	return recordToTask(record), nil
}

func scan(scnr scanner) (*taskRecord, error) {
	var record taskRecord
	if err := scnr.Scan(
		&record.ID,
		&record.UserID,
		&record.Title,
		&record.LeetcodeURL,
		&record.NextNotify,
		&record.NotifyCount,
		&record.IsActive,
		&record.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &record, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return task.ErrNotFound
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		switch pqErr.Code {
		case "23503":
			return task.ErrUserNotFound
		}
	}

	return err
}
