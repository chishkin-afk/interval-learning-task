package taskpg

import "github.com/chishkin-afk/intask/backend/internal/modules/task/domain/task"

func taskToRecord(task *task.Task) *taskRecord {
	return &taskRecord{
		ID:          task.ID(),
		UserID:      task.UserID(),
		Title:       task.Title(),
		LeetcodeURL: task.LeetcodeURL().String(),
		NextNotify:  task.NextNotify(),
		NotifyCount: task.NotifyCount(),
		IsActive:    task.IsActive(),
		CreatedAt:   task.CreatedAt(),
	}
}

func recordToTask(record *taskRecord) *task.Task {
	return task.Restore(
		record.ID,
		record.UserID,
		record.Title,
		task.LeetcodeURL(record.LeetcodeURL),
		record.NextNotify,
		record.NotifyCount,
		record.IsActive,
		record.CreatedAt,
	)
}
