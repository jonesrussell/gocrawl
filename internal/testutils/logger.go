// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements types.Logger for testing
type MockLogger struct {
	mock.Mock
}

// Info implements types.Logger
func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Error implements types.Logger
func (m *MockLogger) Error(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Debug implements types.Logger
func (m *MockLogger) Debug(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Warn implements types.Logger
func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Fatal implements types.Logger
func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

// Printf implements types.Logger
func (m *MockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

// Errorf implements types.Logger
func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

// Sync implements types.Logger
func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockLogger implements types.Logger
var _ types.Logger = (*MockLogger)(nil)
