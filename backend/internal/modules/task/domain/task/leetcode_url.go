package task

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidLeetcodeURL = errors.New("invalid url of leetcode task")

// LeetcodeURL represents a URL of a LeetCode task.
type LeetcodeURL string

// String returns the string representation of the LeetcodeURL.
func (l LeetcodeURL) String() string {
	return string(l)
}

// Validate reports whether the URL is a valid HTTPS URL pointing to
// leetcode.com.
//
// Returns ErrInvalidLeetcodeURL if the URL is invalid.
func (l LeetcodeURL) Validate() error {
	parsed, err := url.Parse(l.String())
	if err != nil {
		return ErrInvalidLeetcodeURL
	}

	if parsed.Scheme != "https" || parsed.Hostname() != "leetcode.com" {
		return ErrInvalidLeetcodeURL
	}

	return nil
}

// Norm returns the normalized representation of the URL.
//
// Currently, normalization trims leading and trailing whitespace.
func (l LeetcodeURL) Norm() LeetcodeURL {
	trimmed := strings.TrimSpace(l.String())
	return LeetcodeURL(trimmed)
}
