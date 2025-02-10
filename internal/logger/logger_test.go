package logger_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestCustomLogger(t *testing.T) {
	// Create a new mock logger
	mockLogger := logger.NewMockCustomLogger()

	// Test Info method
	mockLogger.Info("Info message")
	assert.Contains(t, mockLogger.Messages, "Info message")

	// Test Error method
	mockLogger.Error("Error message")
	assert.Contains(t, mockLogger.Messages, "Error message")

	// Test Debug method
	mockLogger.Debug("Debug message")
	assert.Contains(t, mockLogger.Messages, "Debug message")

	// Test Warn method
	mockLogger.Warn("Warn message")
	assert.Contains(t, mockLogger.Messages, "Warn message")

	// Test Fatalf method
	mockLogger.Fatalf("Fatal message with args: %d", 42)
	assert.Contains(t, mockLogger.Messages, "Fatal message with args: 42")

	// Test Errorf method
	mockLogger.Errorf("Error message with format: %s", "test")
	assert.Contains(t, mockLogger.Messages, "Error message with format: test")
}
