// Package config provides configuration management for the application.
package config

import "testing"

// TestLogger implements Logger for testing
type TestLogger struct {
	T *testing.T
}

// Debug logs a debug message
func (l TestLogger) Debug(msg string, fields ...Field) {
	l.T.Logf("DEBUG: %s %v", msg, fields)
}

// Info logs an info message
func (l TestLogger) Info(msg string, fields ...Field) {
	l.T.Logf("INFO: %s %v", msg, fields)
}

// Warn logs a warning message
func (l TestLogger) Warn(msg string, fields ...Field) {
	l.T.Logf("WARN: %s %v", msg, fields)
}

// Error logs an error message
func (l TestLogger) Error(msg string, fields ...Field) {
	l.T.Logf("ERROR: %s %v", msg, fields)
}

// With returns a new logger with the given fields
func (l TestLogger) With(fields ...Field) Logger {
	return l
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) Logger {
	return TestLogger{T: t}
}
