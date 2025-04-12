// Package testutils provides test utilities for the API package.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx/fxevent"
)

// MockLogger is a mock logger for testing.
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Debug logs a debug message.
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Info logs an info message.
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn logs a warning message.
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error logs an error message.
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal logs a fatal message and exits.
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// With creates a new logger with the given fields.
func (m *MockLogger) With(fields ...any) logger.Interface {
	m.Called(fields)
	return m
}

// NewFxLogger implements logger.Interface
func (m *MockLogger) NewFxLogger() fxevent.Logger {
	args := m.Called()
	if result, ok := args.Get(0).(fxevent.Logger); ok {
		return result
	}
	return &fxevent.NopLogger
}
