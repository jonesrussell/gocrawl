package testing

import "github.com/jonesrussell/gocrawl/internal/logger"

// NopLogger is a no-op implementation of logger.Interface
type NopLogger struct{}

// NewNopLogger creates a new NopLogger
func NewNopLogger() logger.Interface {
	return &NopLogger{}
}

// Debug implements logger.Interface
func (l *NopLogger) Debug(_ string, _ ...any) {}

// Info implements logger.Interface
func (l *NopLogger) Info(_ string, _ ...any) {}

// Warn implements logger.Interface
func (l *NopLogger) Warn(_ string, _ ...any) {}

// Error implements logger.Interface
func (l *NopLogger) Error(_ string, _ ...any) {}

// Errorf implements logger.Interface
func (l *NopLogger) Errorf(_ string, _ ...any) {}

// Printf implements logger.Interface
func (l *NopLogger) Printf(_ string, _ ...any) {}

// Sync implements logger.Interface
func (l *NopLogger) Sync() error {
	return nil
}

// Fatal implements logger.Interface
func (l *NopLogger) Fatal(_ string, _ ...any) {}
