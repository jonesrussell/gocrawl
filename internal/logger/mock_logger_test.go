package logger

import (
	"testing"
)

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()

	t.Run("logging methods", func(t *testing.T) {
		logger.Info("test info")
		logger.Error("test error")
		logger.Debug("test debug")
		logger.Warn("test warn")
		logger.Fatalf("test fatal %s", "arg")
		logger.Errorf("test error %s", "arg")
	})
}
