// Package testutils provides test utilities for the API package.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements logger.Interface for testing.
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger instance.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Info implements logger.Interface.
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements logger.Interface.
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Debug implements logger.Interface.
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements logger.Interface.
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements logger.Interface.
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With implements logger.Interface.
func (m *MockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	if result, ok := args.Get(0).(logger.Interface); ok {
		return result
	}
	// Return a default logger if type assertion fails
	return NewMockLogger()
}
