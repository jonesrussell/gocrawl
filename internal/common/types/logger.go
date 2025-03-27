// Package types provides common interfaces and types used across the application.
package types

// Logger defines the interface for logging operations.
type Logger interface {
	// Info logs an informational message with optional key-value pairs.
	Info(msg string, fields ...any)

	// Error logs an error message with optional key-value pairs.
	Error(msg string, fields ...any)

	// Debug logs a debug message with optional key-value pairs.
	Debug(msg string, fields ...any)

	// Warn logs a warning message with optional key-value pairs.
	Warn(msg string, fields ...any)

	// Fatal logs a fatal message with optional key-value pairs and exits.
	Fatal(msg string, fields ...any)

	// Printf logs a formatted message.
	Printf(format string, args ...any)

	// Errorf logs a formatted error message.
	Errorf(format string, args ...any)

	// Sync flushes any buffered log entries.
	Sync() error
}
