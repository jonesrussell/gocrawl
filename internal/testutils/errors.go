package testutils

import "errors"

var (
	// ErrInvalidResult is returned when a mock result cannot be type asserted
	ErrInvalidResult = errors.New("invalid mock result")
)
