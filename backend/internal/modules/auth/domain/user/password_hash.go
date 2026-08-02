package user

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidPassword = errors.New("invalid password")

// PasswordHash represents a hashed user password.
type PasswordHash string

// String returns the string representation of the password hash.
func (ph PasswordHash) String() string {
	return string(ph)
}

// IsEqual compares the provided plain password with the stored hash.
//
// It returns true if the password matches the hash, otherwise false.
func (ph PasswordHash) IsEqual(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(ph), []byte(password)) == nil
}

// NewPasswordHash creates a new PasswordHash from the provided plain password.
//
// The password is trimmed and validated before hashing.
// The resulting hash is generated using bcrypt with the default cost.
//
// Returns an error if the password length is invalid or the hash generation fails.
func NewPasswordHash(password string) (PasswordHash, error) {
	password = strings.TrimSpace(password)
	n := len([]rune(password))
	if n < 6 || n > 64 {
		return "", fmt.Errorf("%w: len of password must be more than 6 and less than 64",
			ErrInvalidPassword)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrInvalidPassword
	}

	return PasswordHash(hash), nil
}
