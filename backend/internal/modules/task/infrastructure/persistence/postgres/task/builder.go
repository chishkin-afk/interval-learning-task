package taskpg

import "github.com/google/uuid"

func (tr *taskRepository) buildInsertQuery(record *taskRecord) (string, []any, error) {
	return tr.sb.Insert("tasks").Columns(
		taskColumns...,
	).Values(
		record.ID,
		record.UserID,
		record.Title,
		record.LeetcodeURL,
		record.NextNotify,
		record.NotifyCount,
		record.IsActive,
		record.CreatedAt,
	).ToSql()
}

func (tr *taskRepository) buildSelectQuery(id uuid.UUID) (string, []any, error) {
	return tr.sb.Select(taskColumns...).From("tasks").
		Where("id = ?", id).ToSql()
}

func (tr *taskRepository) buildListAllQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	page = max(page, 1)
	offset := limit * (page - 1)
	return tr.sb.Select(taskColumns...).From("tasks").
		Limit(uint64(limit)).Offset(uint64(offset)).
		OrderBy("created_at DESC").Where("user_id = ?", userID).ToSql()
}

func (tr *taskRepository) buildCountAllQuery() (string, []any, error) {
	return tr.sb.Select("count(*)").From("tasks").ToSql()
}

func (tr *taskRepository) buildListByNotificationQuery() (string, []any, error) {
	return tr.sb.Select(taskColumns...).From("tasks").
		Limit(100).
		Where("next_notify <= NOW() AND is_active = TRUE").OrderBy("next_notify ASC").ToSql()
}

func (tr *taskRepository) buildCountByNotificationQuery() (string, []any, error) {
	return tr.sb.Select("count(*)").From("tasks").
		Where("next_notify <= NOW() AND is_active = TRUE").ToSql()
}

func (tr *taskRepository) buildUpdateQuery(record *taskRecord) (string, []any, error) {
	return tr.sb.Update("tasks").SetMap(map[string]any{
		"is_active":    record.IsActive,
		"notify_count": record.NotifyCount,
		"next_notify":  record.NextNotify,
		"title":        record.Title,
		"leetcode_url": record.LeetcodeURL,
	}).Where("id = ?", record.ID).ToSql()
}

func (tr *taskRepository) buildSelectForUpdateQuery(id uuid.UUID) (string, []any, error) {
	return tr.sb.Select(taskColumns...).From("tasks").
		Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

func (tr *taskRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return tr.sb.Delete("tasks").Where("id = ?", id).ToSql()
}
