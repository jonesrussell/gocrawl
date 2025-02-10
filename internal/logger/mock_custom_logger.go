package logger

import (
	"fmt"
	"sync"
)

// MockCustomLogger is a mock implementation of CustomLogger
type MockCustomLogger struct {
	mu       sync.Mutex
	calls    map[string]int
	messages []string
}

// NewMockCustomLogger creates a new instance of MockCustomLogger
func NewMockCustomLogger() *MockCustomLogger {
	return &MockCustomLogger{
		calls:    make(map[string]int),
		messages: make([]string, 0),
	}
}

func (m *MockCustomLogger) recordCall(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[method]++
}

func (m *MockCustomLogger) recordMessage(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

// Implement the methods of CustomLogger
func (m *MockCustomLogger) Info(msg string, fields ...interface{}) {
	m.recordCall("Info")
	m.recordMessage(msg)
}

func (m *MockCustomLogger) Error(msg string, fields ...interface{}) {
	m.recordCall("Error")
	m.recordMessage(msg)
}

func (m *MockCustomLogger) Debug(msg string, fields ...interface{}) {
	m.recordCall("Debug")
	m.recordMessage(msg)
}

func (m *MockCustomLogger) Warn(msg string, fields ...interface{}) {
	m.recordCall("Warn")
	m.recordMessage(msg)
}

// Implement Fatalf method
func (m *MockCustomLogger) Fatalf(msg string, args ...interface{}) {
	m.recordCall("Fatalf")
	m.recordMessage(fmt.Sprintf(msg, args...)) // Format the message correctly
}

// Implement Errorf method
func (m *MockCustomLogger) Errorf(format string, args ...interface{}) {
	m.recordCall("Errorf")
	m.recordMessage(fmt.Sprintf(format, args...)) // Format the message correctly
}

func (m *MockCustomLogger) GetCalls(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[method]
}

func (m *MockCustomLogger) GetMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.messages...)
}
