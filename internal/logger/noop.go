package logger

import "go.uber.org/fx/fxevent"

// NoOpLogger is a logger that does nothing.
type NoOpLogger struct{}

// NewNoOpLogger creates a new no-op logger.
func NewNoOpLogger() Interface {
	return &NoOpLogger{}
}

// Debug logs a debug message.
func (l *NoOpLogger) Debug(msg string, fields ...any) {}

// Info logs an info message.
func (l *NoOpLogger) Info(msg string, fields ...any) {}

// Warn logs a warning message.
func (l *NoOpLogger) Warn(msg string, fields ...any) {}

// Error logs an error message.
func (l *NoOpLogger) Error(msg string, fields ...any) {}

// Fatal logs a fatal message and exits.
func (l *NoOpLogger) Fatal(msg string, fields ...any) {}

// With creates a new logger with the given fields.
func (l *NoOpLogger) With(fields ...any) Interface {
	return l
}

// NewFxLogger implements Interface
func (l *NoOpLogger) NewFxLogger() fxevent.Logger {
	return &fxevent.NopLogger
}
