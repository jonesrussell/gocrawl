package test

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements logger.Interface for testing
type MockLogger struct {
	mock.Mock
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Info logs an info message
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Error logs an error message
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Fatal logs a fatal message
func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// With returns a new logger with the given fields
func (m *MockLogger) With(fields ...interface{}) logger.Interface {
	args := m.Called(fields)
	if args.Get(0) == nil {
		return m
	}
	return args.Get(0).(logger.Interface)
}
