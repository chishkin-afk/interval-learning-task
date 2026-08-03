package task

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type UpdateFunc func(ctx context.Context, task *Task) error

var (
	ErrNotFound     = errors.New("task not found")
	ErrUserNotFound = errors.New("user of task not found")
)

type TaskPersistenceRepository interface {
	Save(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	ListAll(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Task, int64, error)
	ListByNotification(ctx context.Context, page, limit uint32) ([]*Task, int64, error)
	Update(ctx context.Context, id uuid.UUID, updFunc UpdateFunc) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
