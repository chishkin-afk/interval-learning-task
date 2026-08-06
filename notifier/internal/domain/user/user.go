package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user with optional Telegram notification binding.
//
// User fields are unexported; access them via getter methods. Mutation is
// only possible through domain methods (e.g. EnableTG), preserving invariants.
type User struct {
	id        uuid.UUID
	email     string
	tgChatID  int64
	tgEnabled bool
	createdAt time.Time
	updatedAt time.Time
}

// Restore reconstructs a User from persisted state (database, external API).
// It does not validate invariants — callers must ensure the data originated
// from a trusted source that previously enforced them.
//
// For creating brand-new users, use New instead.
func Restore(
	id uuid.UUID,
	email string,
	tgChatID int64,
	tgEnabled bool,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:        id,
		email:     email,
		tgChatID:  tgChatID,
		tgEnabled: tgEnabled,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() string {
	return u.email
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
