package logger

import "go.uber.org/fx/fxevent"

// NoOpLogger is a no-op implementation of Interface
type NoOpLogger struct{}

// NewNoOp creates a new no-op logger
func NewNoOp() Interface {
	return &NoOpLogger{}
}

// Debug implements Interface
func (l *NoOpLogger) Debug(msg string, fields ...any) {}

// Info implements Interface
func (l *NoOpLogger) Info(msg string, fields ...any) {}

// Warn implements Interface
func (l *NoOpLogger) Warn(msg string, fields ...any) {}

// Error implements Interface
func (l *NoOpLogger) Error(msg string, fields ...any) {}

// Fatal implements Interface
func (l *NoOpLogger) Fatal(msg string, fields ...any) {}

// With implements Interface
func (l *NoOpLogger) With(fields ...any) Interface {
	return l
}

// NewFxLogger implements Interface
func (l *NoOpLogger) NewFxLogger() fxevent.Logger {
	return &fxevent.NopLogger
}
