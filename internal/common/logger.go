package common

import (
	"github.com/jonesrussell/gocrawl/internal/common/types"
)

// NewNoOpLogger creates a new no-op logger
func NewNoOpLogger() types.Logger {
	return &noOpLogger{}
}

// noOpLogger is a no-op implementation of types.Logger
type noOpLogger struct{}

func (l *noOpLogger) Info(msg string, fields ...any)    {}
func (l *noOpLogger) Error(msg string, fields ...any)   {}
func (l *noOpLogger) Debug(msg string, fields ...any)   {}
func (l *noOpLogger) Warn(msg string, fields ...any)    {}
func (l *noOpLogger) Fatal(msg string, fields ...any)   {}
func (l *noOpLogger) Printf(format string, args ...any) {}
func (l *noOpLogger) Errorf(format string, args ...any) {}
func (l *noOpLogger) Sync() error                       { return nil }
