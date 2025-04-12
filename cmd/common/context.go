// Package common provides shared utilities for command implementations.
package common

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// LoggerKey is the key used to store the logger in the context
	LoggerKey contextKey = "logger"
)
