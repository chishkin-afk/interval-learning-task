package user

import (
	"context"
	"errors"
)

type UpdateFunc func(ctx context.Context, user *User) error

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)
