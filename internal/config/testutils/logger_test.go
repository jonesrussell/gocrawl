package testutils_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

// TestNewTestLogger verifies the test logger implementation
func TestNewTestLogger(t *testing.T) {
	logger := testutils.NewTestLogger(t)

	// Test Info logging
	logger.Info("test info message")
	logger.Info("test info with fields", config.Field{Key: "test", Value: "value"})

	// Test Warn logging
	logger.Warn("test warn message")
	logger.Warn("test warn with fields", config.Field{Key: "test", Value: "value"})
}
