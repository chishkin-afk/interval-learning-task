package responses

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	LeetcodeURL string    `json:"leetcode_url"`
	NextNotify  time.Time `json:"next_notify"`
	NotifyCount int8      `json:"notify_count"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListTask struct {
	List        []*Task `json:"list"`
	TotalPages  int64   `json:"total_pages"`
	HasNextPage bool    `json:"has_next_page"`
	HasPrevPage bool    `json:"has_prev_page"`
}
