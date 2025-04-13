// Package logger provides logging functionality for the application.
package logger

import "errors"

// Common errors returned by the logger package.
var (
	// ErrInvalidLevel is returned when an invalid logging level is provided.
	ErrInvalidLevel = errors.New("invalid logging level")
	// ErrInvalidEncoding is returned when an invalid log encoding format is provided.
	ErrInvalidEncoding = errors.New("invalid log encoding format")
	// ErrInvalidOutputPath is returned when an invalid output path is provided.
	ErrInvalidOutputPath = errors.New("invalid output path")
	// ErrInvalidErrorOutputPath is returned when an invalid error output path is provided.
	ErrInvalidErrorOutputPath = errors.New("invalid error output path")
	// ErrInvalidFields is returned when invalid fields are provided to a logging method.
	ErrInvalidFields = errors.New("invalid fields: must be key-value pairs")
)
