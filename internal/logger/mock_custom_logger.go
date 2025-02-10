package logger

import (
	"fmt"
)

// MockCustomLogger is a mock implementation of CustomLogger
type MockCustomLogger struct {
	Messages []string
}

// NewMockCustomLogger creates a new instance of MockCustomLogger
func NewMockCustomLogger() *MockCustomLogger {
	return &MockCustomLogger{
		Messages: make([]string, 0),
	}
}

// Implement the methods of CustomLogger
func (m *MockCustomLogger) Info(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Error(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Debug(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Warn(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

// Implement Fatalf method
func (m *MockCustomLogger) Fatalf(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(msg, args...)) // Format the message correctly
}

// Implement Errorf method
func (m *MockCustomLogger) Errorf(format string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...)) // Format the message correctly
}
