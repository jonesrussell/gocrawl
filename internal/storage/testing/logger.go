package testing

import "github.com/jonesrussell/gocrawl/internal/logger"

// NopLogger is a no-op implementation of logger.Interface
type NopLogger struct{}

// NewNopLogger creates a new NopLogger
func NewNopLogger() logger.Interface {
	return &NopLogger{}
}

// Debug implements logger.Interface
func (l *NopLogger) Debug(msg string, args ...interface{}) {}

// Info implements logger.Interface
func (l *NopLogger) Info(msg string, args ...interface{}) {}

// Warn implements logger.Interface
func (l *NopLogger) Warn(msg string, args ...interface{}) {}

// Error implements logger.Interface
func (l *NopLogger) Error(msg string, args ...interface{}) {}

// Errorf implements logger.Interface
func (l *NopLogger) Errorf(format string, args ...interface{}) {}

// Printf implements logger.Interface
func (l *NopLogger) Printf(format string, args ...interface{}) {}

// Sync implements logger.Interface
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal implements logger.Interface
func (l *NopLogger) Fatal(msg string, args ...interface{}) {}
