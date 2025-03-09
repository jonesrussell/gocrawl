package testing

import "github.com/jonesrussell/gocrawl/internal/logger"

// NopLogger is a no-op implementation of logger.Interface
type NopLogger struct{}

// NewNopLogger creates a new NopLogger
func NewNopLogger() logger.Interface {
	return &NopLogger{}
}

// Debug implements logger.Interface
func (l *NopLogger) Debug(_ string, _ ...interface{}) {}

// Info implements logger.Interface
func (l *NopLogger) Info(_ string, _ ...interface{}) {}

// Warn implements logger.Interface
func (l *NopLogger) Warn(_ string, _ ...interface{}) {}

// Error implements logger.Interface
func (l *NopLogger) Error(_ string, _ ...interface{}) {}

// Errorf implements logger.Interface
func (l *NopLogger) Errorf(_ string, _ ...interface{}) {}

// Printf implements logger.Interface
func (l *NopLogger) Printf(_ string, _ ...interface{}) {}

// Sync implements logger.Interface
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal implements logger.Interface
func (l *NopLogger) Fatal(_ string, _ ...interface{}) {}
