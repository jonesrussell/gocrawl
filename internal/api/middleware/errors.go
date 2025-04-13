package middleware

import "errors"

// Error definitions for security middleware
var (
	// ErrMissingAPIKey is returned when the API key is missing
	ErrMissingAPIKey = errors.New("missing API key")

	// ErrInvalidAPIKey is returned when the API key is invalid
	ErrInvalidAPIKey = errors.New("invalid API key")

	// ErrRateLimitExceeded is returned when the rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrOriginNotAllowed is returned when the origin is not allowed
	ErrOriginNotAllowed = errors.New("origin not allowed")
)
