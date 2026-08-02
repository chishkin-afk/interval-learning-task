package userpg

import "github.com/chishkin-afk/intask/backend/internal/modules/auth/domain/user"

func userToRecord(user *user.User) *userRecord {
	return &userRecord{
		ID:           user.ID(),
		TgChatID:     user.TgChatID(),
		TgEnabled:    user.TgEnabled(),
		Email:        user.Email().String(),
		PasswordHash: user.PasswordHash().String(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}

func recordToUser(record *userRecord) *user.User {
	return user.Restore(
		record.ID,
		record.TgChatID,
		record.TgEnabled,
		user.Email(record.Email),
		user.PasswordHash(record.PasswordHash),
		record.CreatedAt,
		record.UpdatedAt,
	)
}
