package task

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyUserID  = errors.New("user id is empty")
	ErrInvalidTitle = errors.New("invalid title of task")

	nextSchedule = []time.Duration{
		24 * time.Hour,
		3 * 24 * time.Hour,
		7 * 24 * time.Hour,
		30 * 24 * time.Hour,
	}
)

// Task represents a LeetCode task assigned to a user.
//
// A task schedules reminder notifications according to the predefined
// notification schedule. Once all scheduled notifications have been sent,
// the task becomes inactive.
type Task struct {
	id          uuid.UUID
	userID      uuid.UUID
	title       string
	leetcodeURL LeetcodeURL
	nextNotify  time.Time
	notifyCount int8
	isActive    bool
	createdAt   time.Time
}

// New creates a new task for the specified user.
//
// The task is initialized as active and the first notification is scheduled
// 24 hours after creation.
func New(
	userID uuid.UUID,
	title string,
	leetcodeURL LeetcodeURL,
) (*Task, error) {
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	leetcodeURL = leetcodeURL.Norm()
	if err := leetcodeURL.Validate(); err != nil {
		return nil, err
	}

	title, err := validateTitle(title)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Task{
		id:          uuid.New(),
		userID:      userID,
		title:       title,
		leetcodeURL: leetcodeURL,
		nextNotify:  now.Add(nextSchedule[0]),
		notifyCount: 0,
		isActive:    true,
		createdAt:   now,
	}, nil
}

func Restore(
	id uuid.UUID,
	userID uuid.UUID,
	title string,
	leetcodeURL LeetcodeURL,
	nextNotify time.Time,
	notifyCount int8,
	isActive bool,
	createdAt time.Time,
) *Task {
	return &Task{
		id:          id,
		userID:      userID,
		title:       title,
		leetcodeURL: leetcodeURL,
		nextNotify:  nextNotify,
		notifyCount: notifyCount,
		isActive:    isActive,
		createdAt:   createdAt,
	}
}

// Notify marks the current notification as sent and schedules the next one.
//
// If all scheduled notifications have been sent, the task is marked as
// inactive. Calling Notify on an inactive task has no effect.
func (t *Task) Notify() {
	if !t.isActive {
		return
	}

	t.notifyCount++
	if int(t.notifyCount) >= len(nextSchedule) {
		t.isActive = false
		return
	}

	t.nextNotify = t.nextNotify.Add(nextSchedule[int(t.notifyCount)])
}

// ChangeTitle updates the task title.
//
// Returns ErrInvalidTitle if the provided title is invalid.
func (t *Task) ChangeTitle(title string) error {
	title, err := validateTitle(title)
	if err != nil {
		return err
	}

	t.title = title

	return nil
}

// ChangeLeetcodeURL updates the task's LeetCode URL.
//
// Returns ErrInvalidLeetcodeURL if the provided URL is invalid.
func (t *Task) ChangeLeetcodeURL(leetcodeURL LeetcodeURL) error {
	leetcodeURL = leetcodeURL.Norm()
	if err := leetcodeURL.Validate(); err != nil {
		return err
	}

	t.leetcodeURL = leetcodeURL

	return nil
}

func (t *Task) ID() uuid.UUID {
	return t.id
}

func (t *Task) UserID() uuid.UUID {
	return t.userID
}

func (t *Task) Title() string {
	return t.title
}

func (t *Task) LeetcodeURL() LeetcodeURL {
	return t.leetcodeURL
}

func (t *Task) NextNotify() time.Time {
	return t.nextNotify
}

func (t *Task) NotifyCount() int8 {
	return t.notifyCount
}

func (t *Task) IsActive() bool {
	return t.isActive
}

func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)

	n := len([]rune(title))
	if n < 3 || n > 128 {
		return "", ErrInvalidTitle
	}

	return title, nil
}
