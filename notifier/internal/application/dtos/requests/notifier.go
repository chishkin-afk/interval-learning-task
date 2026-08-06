package requests

import "github.com/google/uuid"

type SendMsg struct {
	UserID uuid.UUID `json:"user_id"`
	Msg    string    `json:"msg"`
}
