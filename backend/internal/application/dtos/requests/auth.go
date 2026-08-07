package requests

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUser struct {
	TgChatID  *int64 `json:"tg_chat_id"`
	TgEnabled *bool  `json:"telegram_enabled"`
}

type BindTg struct {
	Code   int `json:"code"`
	ChatID int `json:"chat_id"`
}
