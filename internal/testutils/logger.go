// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements logger.Interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockLogger implements logger.Interface
var _ logger.Interface = (*MockLogger)(nil)
