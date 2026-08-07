package backend

import (
	"time"

	"github.com/google/uuid"
)

type userRecord struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	TgChatID  int64     `json:"tg_chat_id"`
	TgEnabled bool      `json:"telegram_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
