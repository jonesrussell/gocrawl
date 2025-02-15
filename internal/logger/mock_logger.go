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
	m.Called(msg)
	m.Messages = append(m.Messages, msg)
}

func (m *MockLogger) Debug(msg string, _ ...interface{}) {
	m.Called(msg)
	m.Messages = append(m.Messages, msg)
}

// Implement Warn method
func (m *MockLogger) Warn(msg string, _ ...interface{}) {
	m.Called(msg)
	m.Messages = append(m.Messages, msg)
}

// Implement Fatal method
func (m *MockLogger) Fatal(msg string, _ ...interface{}) {
	m.Called(msg)
	m.Messages = append(m.Messages, msg)
}

// Implement Fatalf method
func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(fmt.Sprintf(format, args...))
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
}

// Implement Errorf method
func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(fmt.Sprintf(format, args...))
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
}

// Implement Sync method
func (m *MockLogger) Sync() error {
	// Mock implementation, just return nil
	return nil
}

// GetMessages returns the logged messages
func (m *MockLogger) GetMessages() []string {
	return m.Messages
}

// MockCustomLogger is a mock implementation of the logger.Interface
type MockCustomLogger struct {
	mock.Mock
	Messages []string // Store messages for testing
}

// NewMockCustomLogger creates a new instance of MockCustomLogger
func NewMockCustomLogger() *MockCustomLogger {
	return &MockCustomLogger{}
}

// Implement the logger.Interface methods
func (m *MockCustomLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Errorf(format string, args ...interface{}) {
	m.Called(fmt.Sprintf(format, args...))
	m.Messages = append(m.Messages, fmt.Sprintf(format, args...))
}

// Implement Sync method
func (m *MockCustomLogger) Sync() error {
	// Mock implementation, just return nil
	return nil
}

// GetMessages returns the logged messages
func (m *MockCustomLogger) GetMessages() []string {
	return m.Messages
}
