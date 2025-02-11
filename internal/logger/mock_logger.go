package logger

import (
	"fmt"

	"github.com/stretchr/testify/mock"
)

// MockLogger is a simple mock for the logger
type MockLogger struct {
	mock.Mock
	Messages []string
}

// NewMockLogger creates a new instance of MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Messages: make([]string, 0), // Initialize the Messages slice
	}
}

// Implement the same methods as CustomLogger
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Error(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Warn(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

// Implement Fatalf method
func (m *MockLogger) Fatalf(msg string, _ ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

// Implement Errorf method
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
}

// Add any other methods that CustomLogger has

// MockCustomLogger is a mock implementation of the logger.Interface
type MockCustomLogger struct {
	mock.Mock
}

func (m *MockCustomLogger) Fatal(args ...interface{}) {
	m.Called(args)
}

func (m *MockCustomLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Implement Debug method
func (m *MockCustomLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// Implement Errorf method
func (m *MockCustomLogger) Errorf(format string, args ...interface{}) {
	m.Called(fmt.Sprintf(format, args...))
}

// Implement other methods as needed...

// NewMockCustomLogger creates a new instance of MockCustomLogger
func NewMockCustomLogger() *MockCustomLogger {
	return &MockCustomLogger{}
}
