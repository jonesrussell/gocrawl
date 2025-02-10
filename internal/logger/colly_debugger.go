package logger

import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// CollyDebugger implements colly.Debugger interface
type CollyDebugger struct {
	Logger Interface
}

// Init implements colly.Debugger
func (d *CollyDebugger) Init() error {
	return nil
}

// Event implements colly.Debugger
func (d *CollyDebugger) Event(e *debug.Event) {
	if d.Logger == nil {
		return
	}

	d.Logger.Debug("Colly event",
		"type", e.Type,
		"requestID", e.RequestID,
		"collectorID", e.CollectorID,
	)
}

// OnRequest logs the request details
func (d *CollyDebugger) OnRequest(e *colly.Request) {
	d.Logger.Debug(
		"Request",
		"url", e.URL.String(),
		"method", e.Method,
		"headers", e.Headers,
	)
}

// OnResponse logs the response details
func (d *CollyDebugger) OnResponse(e *colly.Response) {
	d.Logger.Info(
		"Response",
		"url", e.Request.URL.String(),
		"status", e.StatusCode,
		"headers", e.Headers,
	)
}

// OnError logs the error details
func (d *CollyDebugger) OnError(e *colly.Response, err error) {
	d.Logger.Error(
		"Error",
		"url", e.Request.URL.String(),
		"status", e.StatusCode,
		"error", err.Error(),
	)
}

// OnEvent logs the event details
func (d *CollyDebugger) OnEvent(e *debug.Event) {
	d.Logger.Info(
		"Event",
		"type", e.Type,
		"requestID", e.RequestID,
		"collectorID", e.CollectorID,
	)
}
