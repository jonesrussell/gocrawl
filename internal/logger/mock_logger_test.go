package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

func TestMockLogger(t *testing.T) {
	mockLogger := logger.NewMockLogger()

	t.Run("logging methods", func(_ *testing.T) {
		// Set expectations for the Info method
		mockLogger.On("Info", "test info", mock.Anything).Return()
		mockLogger.On("Error", "test error", mock.Anything).Return()
		mockLogger.On("Debug", "test debug", mock.Anything).Return()
		mockLogger.On("Warn", "test warn", mock.Anything).Return()
		mockLogger.On("Fatal", "test fatal %s", mock.Anything).Return()
		mockLogger.On("Error", "test error %s", mock.Anything).Return()
		mockLogger.On("Fatalf", "test fatal %s", mock.Anything).Return()

		mockLogger.Info("test info", "key", "value")
		mockLogger.Error("test error", "key", "value")
		mockLogger.Debug("test debug", "key", "value")
		mockLogger.Warn("test warn", "key", "value")
		mockLogger.Fatalf("test fatal %s", "arg")
		mockLogger.Errorf("test error %s", "arg")

		// Test Sync
		if syncErr := mockLogger.Sync(); syncErr != nil {
			// Ignore sync errors as they're expected when writing to console
			t.Log("Sync() error (expected):", syncErr)
		}

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})
}
