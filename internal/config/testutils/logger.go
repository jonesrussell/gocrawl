package testutils

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
)

// TestLogger implements config.Logger for testing
type TestLogger struct {
	T *testing.T
}

func (l TestLogger) Debug(msg string, fields ...config.Field) {
	l.T.Logf("DEBUG: %s %v", msg, fields)
}

func (l TestLogger) Info(msg string, fields ...config.Field) {
	l.T.Logf("INFO: %s %v", msg, fields)
}

func (l TestLogger) Warn(msg string, fields ...config.Field) {
	l.T.Logf("WARN: %s %v", msg, fields)
}

func (l TestLogger) Error(msg string, fields ...config.Field) {
	l.T.Logf("ERROR: %s %v", msg, fields)
}

func (l TestLogger) With(fields ...config.Field) config.Logger {
	return l
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) config.Logger {
	return TestLogger{T: t}
}
