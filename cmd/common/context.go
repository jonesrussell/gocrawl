// Package common provides shared utilities for command implementations.
package common

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// LoggerKey is the key used to store the logger in the context
var LoggerKey contextKey = "logger"

// ConfigKey is the context key for the configuration
var ConfigKey contextKey = "config"
