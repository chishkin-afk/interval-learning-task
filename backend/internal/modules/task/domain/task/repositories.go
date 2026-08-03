package task

import (
	"context"
	"errors"
)

type UpdateFunc func(ctx context.Context, task *Task) error

var (
	ErrNotFound     = errors.New("task not found")
	ErrUserNotFound = errors.New("user of task not found")
)
