package logger

import (
	"time"
)

// NoOpLogger is a logger that does nothing.
type NoOpLogger struct{}

// NewNoOpLogger creates a new no-op logger.
func NewNoOpLogger() Interface {
	return &NoOpLogger{}
}

// NewNoOp creates a new no-op logger instance.
func NewNoOp() Interface {
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

// WithUser adds a user ID to the logger.
func (l *NoOpLogger) WithUser(userID string) Interface {
	return l
}

// WithRequestID adds a request ID to the logger.
func (l *NoOpLogger) WithRequestID(requestID string) Interface {
	return l
}

// WithTraceID adds a trace ID to the logger.
func (l *NoOpLogger) WithTraceID(traceID string) Interface {
	return l
}

// WithSpanID adds a span ID to the logger.
func (l *NoOpLogger) WithSpanID(spanID string) Interface {
	return l
}

// WithDuration adds a duration to the logger.
func (l *NoOpLogger) WithDuration(duration time.Duration) Interface {
	return l
}

// WithError adds an error to the logger.
func (l *NoOpLogger) WithError(err error) Interface {
	return l
}

// WithComponent adds a component name to the logger.
func (l *NoOpLogger) WithComponent(component string) Interface {
	return l
}

// WithVersion adds a version to the logger.
func (l *NoOpLogger) WithVersion(version string) Interface {
	return l
}

// WithEnvironment adds an environment to the logger.
func (l *NoOpLogger) WithEnvironment(env string) Interface {
	return l
}
