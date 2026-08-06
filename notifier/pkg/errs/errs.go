package errs

import "errors"

var (
	ErrInternalServer = errors.New("internal server error")
	ErrTgDisabled     = errors.New("tg of user is disabled")
	ErrInvalidMessage = errors.New("invalid message")
)
