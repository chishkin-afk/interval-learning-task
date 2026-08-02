package responses

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	Token string        `json:"token"`
	TTL   time.Duration `json:"ttl"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	TgChatID  int64     `json:"tg_chat_id"`
	TgEnabled bool      `json:"telegram_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
