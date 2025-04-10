// Package config provides configuration management for the GoCrawl application.
package config

// Logger defines the interface for logging operations.
type Logger interface {
	// Debug logs a message at debug level
	Debug(msg string, fields ...Field)
	// Info logs a message at info level
	Info(msg string, fields ...Field)
	// Warn logs a message at warn level
	Warn(msg string, fields ...Field)
	// Error logs a message at error level
	Error(msg string, fields ...Field)
	// With returns a new logger with the given fields
	With(fields ...Field) Logger
}

// Field represents a single logging field.
type Field struct {
	// Key is the field name
	Key string
	// Value is the field value
	Value interface{}
}
