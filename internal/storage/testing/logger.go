package testing

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// NopLogger is a no-op implementation of logger.Interface
type NopLogger struct{}

// NewNopLogger creates a new no-op logger.
func NewNopLogger() logger.Interface {
	return &NopLogger{}
}

// Debug implements logger.Interface
func (l *NopLogger) Debug(msg string, fields ...any) {}

// Info implements logger.Interface
func (l *NopLogger) Info(msg string, fields ...any) {}

// Warn implements logger.Interface
func (l *NopLogger) Warn(msg string, fields ...any) {}

// Error implements logger.Interface
func (l *NopLogger) Error(msg string, fields ...any) {}

// Errorf implements logger.Interface
func (l *NopLogger) Errorf(format string, args ...any) {}

// Printf implements logger.Interface
func (l *NopLogger) Printf(format string, args ...any) {}

// Sync implements logger.Interface
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal implements logger.Interface
func (l *NopLogger) Fatal(msg string, fields ...any) {}

// With implements logger.Interface
func (l *NopLogger) With(fields ...any) logger.Interface {
	return l
}
