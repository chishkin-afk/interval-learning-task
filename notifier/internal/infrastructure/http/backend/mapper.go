package backend

import "github.com/chishkin/intask/notifier/internal/domain/user"

func recordToUser(record *userRecord) *user.User {
	return user.Restore(
		record.ID,
		record.Email,
		record.TgChatID,
		record.TgEnabled,
		record.CreatedAt,
		record.UpdatedAt,
	)
}
