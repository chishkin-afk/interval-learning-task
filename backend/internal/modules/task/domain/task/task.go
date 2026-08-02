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
)

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
		nextNotify:  now.Add(24 * time.Hour),
		notifyCount: 1,
		isActive:    true,
		createdAt:   now,
	}, nil
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
