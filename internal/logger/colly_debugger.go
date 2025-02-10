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
		d.logger.Field("url", e.URL),
		d.logger.Field("method", e.Method),
		d.logger.Field("headers", e.Headers),
	)
}

// OnResponse logs the response details
func (d *CollyDebugger) OnResponse(e *colly.Response) {
	d.logger.Info(
		"Response",
		d.logger.Field("url", e.Request.URL),
		d.logger.Field("status", e.StatusCode),
		d.logger.Field("headers", e.Headers),
	)
}

// OnError logs the error details
func (d *CollyDebugger) OnError(e *colly.Response, err error) {
	d.logger.Error(
		"Error",
		d.logger.Field("url", e.Request.URL),
		d.logger.Field("status", e.StatusCode),
		d.logger.Field("error", err.Error()),
	)
}

// OnEvent logs the event details
func (d *CollyDebugger) OnEvent(e *debug.Event) {
	d.logger.Info(
		"Event",
		d.logger.Field("type", e.Type),
		d.logger.Field("request_id", e.RequestID),
		d.logger.Field("collector_id", e.CollectorID),
	)
}

// Implement the Event method to satisfy the debug.Debugger interface
func (d *CollyDebugger) Event(e *debug.Event) {
	d.logger.Info(
		"Event",
		d.logger.Field("type", e.Type),
		d.logger.Field("request_id", e.RequestID),
		d.logger.Field("collector_id", e.CollectorID),
	)
}
