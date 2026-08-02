package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTgAlreadyDisabled = errors.New("tg is already disabled")
	ErrTgAlreadyEnabled  = errors.New("tg is already enabled")
)

// User represents a user entity in the system.
//
// User contains user identity information, authentication data,
// and Telegram notification settings.
type User struct {
	id           uuid.UUID
	tgChatID     int64
	tgEnabled    bool
	email        Email
	passwordHash PasswordHash
	createdAt    time.Time
	updatedAt    time.Time
}

// New creates a new User entity with the provided email and password.
//
// The email is normalized and validated before creating the entity.
// The provided password is hashed using bcrypt.
//
// Returns an error if the email or password is invalid.
func New(email Email, password string) (*User, error) {
	email = email.Norm()
	if err := email.Validate(); err != nil {
		return nil, err
	}

	passwordHash, err := NewPasswordHash(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		id:           uuid.New(),
		email:        email,
		passwordHash: passwordHash,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ChangeTgChatID updates the Telegram chat ID associated with the user.
//
// If TgChatID is zero, Telegram integration is disabled.
// The update timestamp is refreshed after the change.
func (u *User) ChangeTgChatID(TgChatID int64) {
	if TgChatID == 0 {
		u.DisableTg()
	}

	u.tgChatID = TgChatID
	u.updatedAt = time.Now().UTC()
}

// DisableTg disables Telegram notifications for the user.
//
// Returns ErrTgAlreadyDisabled if Telegram notifications are already disabled.
func (u *User) DisableTg() error {
	if !u.tgEnabled {
		return ErrTgAlreadyDisabled
	}

	u.tgEnabled = false
	u.updatedAt = time.Now().UTC()

	return nil
}

// EnableTg enables Telegram notifications for the user.
//
// Returns ErrTgAlreadyEnabled if Telegram notifications are already enabled.
func (u *User) EnableTg() error {
	if u.tgEnabled {
		return ErrTgAlreadyEnabled
	}

	u.tgEnabled = true
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) TgChatID() int64 {
	return u.tgChatID
}

func (u *User) TgEnabled() bool {
	return u.tgEnabled
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}
