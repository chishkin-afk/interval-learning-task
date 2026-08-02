package user

import (
	"errors"
	"net/mail"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email")

// Email represents a normalized user email address.
//
// Email is a value object used to encapsulate email validation
// and normalization logic.
type Email string

// Validate checks whether the email address has a valid format.
//
// It returns ErrInvalidEmail if the email cannot be parsed as a valid
// email address.
func (e Email) Validate() error {
	if _, err := mail.ParseAddress(e.String()); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

// String returns the string representation of the email address.z
func (e Email) String() string {
	return string(e)
}

// Norm returns a normalized version of the email address.
//
// Normalization trims leading and trailing spaces and converts
// the email address to lowercase.
func (e Email) Norm() Email {
	email := strings.TrimSpace(e.String())
	return Email(strings.ToLower(email))
}

// NewEmail creates a new Email value object from the provided email string.
//
// The input email is normalized before validation. If the normalized value
// does not represent a valid email address, NewEmail returns an error.
//
// Returned Email is guaranteed to be normalized and valid.
func NewEmail(email string) (Email, error) {
	normalized := Email(email).Norm()

	if err := normalized.Validate(); err != nil {
		return "", err
	}

	return normalized, nil
}
