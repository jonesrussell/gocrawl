package logger

// MockLogger is a simple mock for the logger
type MockLogger struct {
	Messages []string
}

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
