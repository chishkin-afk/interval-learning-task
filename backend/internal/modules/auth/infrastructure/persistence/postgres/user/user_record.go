package userpg

import (
	"time"

	"github.com/google/uuid"
)

type userRecord struct {
	ID           uuid.UUID
	TgChatID     int64
	TgEnabled    bool
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
