// Package interfaces provides shared interfaces used across the GoCrawl application.
package interfaces

// Logger defines the interface for logging operations.
// It provides structured logging capabilities with different log levels and
// support for additional fields in log messages.
type Logger interface {
	// Debug logs a debug message with optional fields.
	// Used for detailed information useful during development.
	Debug(msg string, fields ...any)
	// Error logs an error message with optional fields.
	// Used for error conditions that need immediate attention.
	Error(msg string, fields ...any)
	// Info logs an informational message with optional fields.
	// Used for general operational information.
	Info(msg string, fields ...any)
	// Warn logs a warning message with optional fields.
	// Used for potentially harmful situations.
	Warn(msg string, fields ...any)
	// Fatal logs a fatal message and panics.
	// Used for unrecoverable errors that require immediate termination.
	Fatal(msg string, fields ...any)
	// Printf logs a formatted message.
	// Used for formatted string logging.
	Printf(format string, args ...any)
	// Errorf logs a formatted error message.
	// Used for formatted error string logging.
	Errorf(format string, args ...any)
	// Sync flushes any buffered log entries.
	// Used to ensure all logs are written before shutdown.
	Sync() error
}
