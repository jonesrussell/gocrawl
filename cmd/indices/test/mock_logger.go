package test

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of logger.Interface
type MockLogger struct {
	mock.Mock
}

// Debug implements logger.Interface
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(fields)
}

// Info implements logger.Interface
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(fields)
}

// Warn implements logger.Interface
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(fields)
}

// Error implements logger.Interface
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(fields)
}

// Fatal implements logger.Interface
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(fields)
}

// With implements logger.Interface
func (m *MockLogger) With(fields ...any) logger.Interface {
	args := m.Called(fields)
	iface, _ := args.Get(0).(logger.Interface)
	return iface
}
