package testing

import (
	"github.com/jonesrussell/gocrawl/internal/common"
)

// NopLogger is a no-op implementation of common.Logger
type NopLogger struct{}

// NewNopLogger creates a new no-op logger
func NewNopLogger() common.Logger {
	return &NopLogger{}
}

// Debug implements common.Logger
func (l *NopLogger) Debug(msg string, fields ...any) {}

// Info implements common.Logger
func (l *NopLogger) Info(msg string, fields ...any) {}

// Warn implements common.Logger
func (l *NopLogger) Warn(msg string, fields ...any) {}

// Error implements common.Logger
func (l *NopLogger) Error(msg string, fields ...any) {}

// Errorf implements common.Logger
func (l *NopLogger) Errorf(format string, args ...any) {}

// Printf implements common.Logger
func (l *NopLogger) Printf(format string, args ...any) {}

// Sync implements common.Logger
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal implements common.Logger
func (l *NopLogger) Fatal(msg string, fields ...any) {}
