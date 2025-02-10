package logger

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// CollyDebugger is a custom debugger for Colly
type CollyDebugger struct {
	logger *CustomLogger
}

// NewCollyDebugger creates a new CollyDebugger
func NewCollyDebugger(logger *CustomLogger) *CollyDebugger {
	return &CollyDebugger{logger: logger}
}

// Init initializes the debugger (no specific initialization needed here)
func (d *CollyDebugger) Init() error {
	// You can add any initialization logic if needed
	return nil
}

// OnRequest logs the request details
func (d *CollyDebugger) OnRequest(e *colly.Request) {
	d.logger.Debug(
		"Request",
		e.URL,
		e.Method,
		e.Headers,
	)
}

// OnResponse logs the response details
func (d *CollyDebugger) OnResponse(e *colly.Response) {
	d.logger.Info(
		"Response",
		e.Request.URL,
		e.StatusCode,
		e.Headers,
	)
}

// OnError logs the error details
func (d *CollyDebugger) OnError(e *colly.Response, err error) {
	d.logger.Error(
		"Error",
		e.Request.URL,
		e.StatusCode,
		err.Error(),
	)
}

// OnEvent logs the event details
func (d *CollyDebugger) OnEvent(e *debug.Event) {
	d.logger.Info(
		"Event",
		e.Type,
		e.RequestID,
		e.CollectorID,
	)
}

// Implement the Event method to satisfy the debug.Debugger interface
func (d *CollyDebugger) Event(e *debug.Event) {
	d.logger.Info(
		"Event",
		e.Type,
		e.RequestID,
		e.CollectorID,
	)
}
