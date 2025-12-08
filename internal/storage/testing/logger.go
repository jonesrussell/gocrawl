// Package testing provides a no-op implementation of the logger interface for testing purposes.
package testing

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// NopLogger is a no-op implementation of logger.Interface
type NopLogger struct{}

// NewNopLogger creates a new no-op logger instance
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

// WithUser adds a user ID to the logger.
func (l *NopLogger) WithUser(userID string) logger.Interface {
	return l
}

// WithRequestID adds a request ID to the logger.
func (l *NopLogger) WithRequestID(requestID string) logger.Interface {
	return l
}

// WithTraceID adds a trace ID to the logger.
func (l *NopLogger) WithTraceID(traceID string) logger.Interface {
	return l
}

// WithSpanID adds a span ID to the logger.
func (l *NopLogger) WithSpanID(spanID string) logger.Interface {
	return l
}

// WithDuration adds a duration to the logger.
func (l *NopLogger) WithDuration(duration time.Duration) logger.Interface {
	return l
}

// WithError adds an error to the logger.
func (l *NopLogger) WithError(err error) logger.Interface {
	return l
}

// WithComponent adds a component name to the logger.
func (l *NopLogger) WithComponent(component string) logger.Interface {
	return l
}

// WithVersion adds a version to the logger.
func (l *NopLogger) WithVersion(version string) logger.Interface {
	return l
}

// WithEnvironment adds an environment to the logger.
func (l *NopLogger) WithEnvironment(env string) logger.Interface {
	return l
}
