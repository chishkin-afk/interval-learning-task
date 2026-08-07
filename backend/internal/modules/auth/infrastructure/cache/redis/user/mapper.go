package userredis

import (
	"encoding/json"

	"github.com/chishkin-afk/intask/backend/internal/modules/auth/domain/user"
)

func userToBytes(user *user.User) ([]byte, error) {
	return json.Marshal(userRecord{
		ID:        user.ID(),
		Email:     user.Email().String(),
		TgChatID:  user.TgChatID(),
		TgEnabled: user.TgEnabled(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	})
}

func bytesToUser(bytes []byte) (*user.User, error) {
	var record userRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return user.Restore(
		record.ID,
		record.TgChatID,
		record.TgEnabled,
		user.Email(record.Email),
		"",
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
