// Package logger provides logging functionality for the application.
package logger

// Interface defines the interface for logging operations.
// It provides structured logging capabilities with different log levels and
// support for additional fields in log messages.
type Interface interface {
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

// Params holds the parameters for creating a logger
type Params struct {
	Debug  bool
	Level  string
	AppEnv string
}

// NewNoOp creates a no-op logger that discards all log messages.
// This is useful for testing or when logging is not needed.
func NewNoOp() Interface {
	return &NoOpLogger{}
}

// NoOpLogger implements Interface but discards all log messages.
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, fields ...any)   {}
func (l *NoOpLogger) Error(msg string, fields ...any)   {}
func (l *NoOpLogger) Info(msg string, fields ...any)    {}
func (l *NoOpLogger) Warn(msg string, fields ...any)    {}
func (l *NoOpLogger) Fatal(msg string, fields ...any)   {}
func (l *NoOpLogger) Printf(format string, args ...any) {}
func (l *NoOpLogger) Errorf(format string, args ...any) {}
func (l *NoOpLogger) Sync() error                       { return nil }
