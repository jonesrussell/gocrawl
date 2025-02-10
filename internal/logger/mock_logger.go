package logger

import (
	"fmt"
)

// MockLogger is a simple mock for the logger
type MockLogger struct {
	Messages []string
}

// NewMockLogger creates a new instance of MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: make([]string, 0), // Initialize the Messages slice
	}
}

// Implement the same methods as CustomLogger
func (m *MockLogger) Info(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Error(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Debug(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Warn(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

// Implement Fatalf method
func (m *MockLogger) Fatalf(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

// Implement Errorf method
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
}

// Add any other methods that CustomLogger has
