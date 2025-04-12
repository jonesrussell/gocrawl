// Package logger provides logging functionality for the application.
package logger

import "go.uber.org/fx/fxevent"

// Interface defines the interface for logging operations.
type Interface interface {
	// Debug logs a debug message.
	Debug(msg string, fields ...any)
	// Info logs an info message.
	Info(msg string, fields ...any)
	// Warn logs a warning message.
	Warn(msg string, fields ...any)
	// Error logs an error message.
	Error(msg string, fields ...any)
	// Fatal logs a fatal message and exits.
	Fatal(msg string, fields ...any)
	// With creates a child logger with additional fields.
	With(fields ...any) Interface
	// NewFxLogger creates a new Fx logger.
	NewFxLogger() fxevent.Logger
}
