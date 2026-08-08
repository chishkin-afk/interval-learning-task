package notifier

import "github.com/google/uuid"

type sendMsg struct {
	UserID uuid.UUID `json:"user_id"`
	Msg    string    `json:"msg"`
}
