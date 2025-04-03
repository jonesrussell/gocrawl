// Package logger provides logging functionality for the application.
package logger

// Interface defines the interface for logging operations.
type Interface interface {
	// Debug logs a debug message.
	Debug(msg string, fields ...interface{})
	// Info logs an info message.
	Info(msg string, fields ...interface{})
	// Warn logs a warning message.
	Warn(msg string, fields ...interface{})
	// Error logs an error message.
	Error(msg string, fields ...interface{})
	// Fatal logs a fatal message and exits.
	Fatal(msg string, fields ...interface{})
	// With creates a child logger with additional fields.
	With(fields ...interface{}) Interface
}
