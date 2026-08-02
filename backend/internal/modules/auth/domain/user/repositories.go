package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type UpdateFunc func(ctx context.Context, user *User) error

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)

type UserPersistenceRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Update(ctx context.Context, id uuid.UUID, updFunc UpdateFunc) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
