// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx/fxevent"
)

// MockLogger implements logger.Interface for testing
type MockLogger struct {
	mock.Mock
}

// Info implements logger.Interface
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements logger.Interface
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Debug implements logger.Interface
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements logger.Interface
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements logger.Interface
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With implements logger.Interface
func (m *MockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	if result, ok := args.Get(0).(logger.Interface); ok {
		return result
	}
	// Return a default logger if type assertion fails
	return NewMockLogger()
}

// NewFxLogger implements logger.Interface
func (m *MockLogger) NewFxLogger() fxevent.Logger {
	args := m.Called()
	if result, ok := args.Get(0).(fxevent.Logger); ok {
		return result
	}
	return &fxevent.NopLogger
}

// Ensure MockLogger implements logger.Interface
var _ logger.Interface = (*MockLogger)(nil)

// NewMockLogger creates a new mock logger instance.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}
