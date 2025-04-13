package testing

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx/fxevent"
)

// NopLogger is a no-op logger for testing.
type NopLogger struct{}

// NewNopLogger creates a new no-op logger.
func NewNopLogger() logger.Interface {
	return &NopLogger{}
}

// Debug logs a debug message.
func (l *NopLogger) Debug(msg string, fields ...any) {}

// Info logs an info message.
func (l *NopLogger) Info(msg string, fields ...any) {}

// Warn logs a warning message.
func (l *NopLogger) Warn(msg string, fields ...any) {}

// Error logs an error message.
func (l *NopLogger) Error(msg string, fields ...any) {}

// Errorf implements logger.Interface
func (l *NopLogger) Errorf(format string, args ...any) {}

// Printf implements logger.Interface
func (l *NopLogger) Printf(format string, args ...any) {}

// Sync implements logger.Interface
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal logs a fatal message and exits.
func (l *NopLogger) Fatal(msg string, fields ...any) {}

// With creates a new logger with the given fields.
func (l *NopLogger) With(fields ...any) logger.Interface {
	return l
}

// NewFxLogger implements logger.Interface
func (l *NopLogger) NewFxLogger() fxevent.Logger {
	return &fxevent.NopLogger
}
