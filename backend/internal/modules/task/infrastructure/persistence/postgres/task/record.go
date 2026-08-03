package taskpg

import (
	"time"

	"github.com/google/uuid"
)

type taskRecord struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	LeetcodeURL string
	NextNotify  time.Time
	NotifyCount int8
	IsActive    bool
	CreatedAt   time.Time
}
