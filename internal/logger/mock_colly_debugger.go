package logger

import (
	"strconv"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// MockCollyDebugger is a mock implementation of the CollyDebugger
type MockCollyDebugger struct {
	Messages []string
}

// NewMockCollyDebugger creates a new instance of MockCollyDebugger
func NewMockCollyDebugger() *MockCollyDebugger {
	return &MockCollyDebugger{
		Messages: make([]string, 0),
	}
}

// Init initializes the debugger (no specific initialization needed here)
func (m *MockCollyDebugger) Init() error {
	return nil
}

// OnRequest logs the request details
func (m *MockCollyDebugger) OnRequest(e *colly.Request) {
	m.Messages = append(m.Messages, "Request: "+e.URL.String())
}

// OnResponse logs the response details
func (m *MockCollyDebugger) OnResponse(e *colly.Response) {
	m.Messages = append(m.Messages, "Response: "+e.Request.URL.String()+" Status: "+strconv.Itoa(e.StatusCode))
}

// OnError logs the error details
func (m *MockCollyDebugger) OnError(e *colly.Response, err error) {
	m.Messages = append(m.Messages, "Error: "+e.Request.URL.String()+" Error: "+err.Error())
}

// OnEvent logs the event details
func (m *MockCollyDebugger) OnEvent(e *debug.Event) {
	m.Messages = append(m.Messages, "Event: "+e.Type)
}

// Event implements the debug.Debugger interface
func (m *MockCollyDebugger) Event(e *debug.Event) {
	m.Messages = append(m.Messages, "Event: "+e.Type)
}
