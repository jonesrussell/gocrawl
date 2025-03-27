// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements common.Logger for testing
type MockLogger struct {
	mock.Mock
}

// Info implements common.Logger
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements common.Logger
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Debug implements common.Logger
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements common.Logger
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements common.Logger
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Printf implements common.Logger
func (m *MockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

// Errorf implements common.Logger
func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

// Sync implements common.Logger
func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockLogger implements common.Logger
var _ common.Logger = (*MockLogger)(nil)
