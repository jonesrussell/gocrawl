// Package testutil provides testing utilities for the application.
package testutil

import (
	"testing"
)

// MockLogger is a mock implementation of the logger interface for testing.
type MockLogger struct {
	T *testing.T
}

// NewMockLogger creates a new mock logger for testing.
func NewMockLogger(t *testing.T) *MockLogger {
	return &MockLogger{T: t}
}

// Info logs an info message.
func (l *MockLogger) Info(msg string, fields ...any) {
	l.T.Logf("INFO: %s %v", msg, fields)
}

// Error logs an error message.
func (l *MockLogger) Error(msg string, fields ...any) {
	l.T.Logf("ERROR: %s %v", msg, fields)
}

// Fatal logs a fatal message and fails the test.
func (l *MockLogger) Fatal(msg string, fields ...any) {
	l.T.Fatalf("FATAL: %s %v", msg, fields)
}

// Debug logs a debug message.
func (l *MockLogger) Debug(msg string, fields ...any) {
	l.T.Logf("DEBUG: %s %v", msg, fields)
}

// Warn logs a warning message.
func (l *MockLogger) Warn(msg string, fields ...any) {
	l.T.Logf("WARN: %s %v", msg, fields)
}

// With returns a new logger with the given fields.
func (l *MockLogger) With(fields ...any) *MockLogger {
	l.T.Logf("WITH: %v", fields)
	return l
}

// Printf formats and logs a message.
func (l *MockLogger) Printf(format string, args ...any) {
	l.T.Log("PRINT: " + format)
}
