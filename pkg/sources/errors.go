// Package sources provides source management functionality for the application.
package sources

import "errors"

var (
	// ErrSourceNotFound is returned when a source is not found.
	ErrSourceNotFound = errors.New("source not found")
	// ErrInvalidSource is returned when a source configuration is invalid.
	ErrInvalidSource = errors.New("invalid source configuration")
	// ErrInvalidURL is returned when a source URL is invalid.
	ErrInvalidURL = errors.New("invalid source URL")
	// ErrInvalidRateLimit is returned when a source rate limit is invalid.
	ErrInvalidRateLimit = errors.New("invalid rate limit")
	// ErrInvalidMaxDepth is returned when a source max depth is invalid.
	ErrInvalidMaxDepth = errors.New("invalid max depth")
	// ErrInvalidTime is returned when a source time is invalid.
	ErrInvalidTime = errors.New("invalid time")
)
