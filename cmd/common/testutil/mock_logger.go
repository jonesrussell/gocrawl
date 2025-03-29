// Package testutil provides utilities for testing Cobra commands
package testutil

import "github.com/jonesrussell/gocrawl/internal/common"

// MockLogger implements common.Logger for testing
type MockLogger struct {
	common.Logger
	InfoCalled  bool
	ErrorCalled bool
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Info implements common.Logger
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.InfoCalled = true
}

// Error implements common.Logger
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.ErrorCalled = true
}
