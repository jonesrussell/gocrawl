package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestMockLogger(t *testing.T) {
	logger := logger.NewMockCustomLogger()

	t.Run("logging methods", func(_ *testing.T) {
		logger.Info("test info")
		logger.Error("test error")
		logger.Debug("test debug")
		logger.Warn("test warn")
		logger.Fatal("test fatal %s", "arg")
		logger.Error("test error %s", "arg")
	})
}
