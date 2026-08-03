package errs

import "errors"

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrInternalServer     = errors.New("internal server error")
	ErrNilRequest         = errors.New("request is nil")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotAuthor      = errors.New("user isn't author")

	ErrInvalidPage  = errors.New("invalid page")
	ErrInvalidLimit = errors.New("invalid limit")
)
