package logger

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
func (m *MockCustomLogger) Info(msg string, fields ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Error(msg string, fields ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Debug(msg string, fields ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Warn(msg string, fields ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Fatalf(msg string, args ...interface{}) {
	m.Messages = append(m.Messages, msg)
}

func (m *MockCustomLogger) Errorf(format string, args ...interface{}) {
	m.Messages = append(m.Messages, format)
}
