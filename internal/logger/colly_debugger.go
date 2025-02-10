package logger

import (
	"github.com/gocolly/colly/v2/debug"
)

// CustomDebugger is a custom debugger that uses ZapLogger
type CustomDebugger struct {
	logger Interface
}

// NewCustomDebugger creates a new CustomDebugger
func NewCustomDebugger(logger Interface) *CustomDebugger {
	return &CustomDebugger{logger: logger}
}

// Init initializes the debugger (no specific initialization needed here)
func (d *CustomDebugger) Init() error {
	// You can add any initialization logic if needed
	return nil
}

// OnRequest logs the request details
func (d *CustomDebugger) OnRequest(e *debug.Event) {
	d.logger.Debug("Request", d.logger.Field("request_id", e.RequestID), d.logger.Field("collector_id", e.CollectorID), d.logger.Field("type", e.Type))
}

// OnResponse logs the response details
func (d *CustomDebugger) OnResponse(e *debug.Event) {
	d.logger.Info("Response", d.logger.Field("request_id", e.RequestID), d.logger.Field("collector_id", e.CollectorID), d.logger.Field("type", e.Type))
}

// OnError logs errors
func (d *CustomDebugger) OnError(e *debug.Event, err error) {
	d.logger.Error("Error", d.logger.Field("request_id", e.RequestID), d.logger.Field("collector_id", e.CollectorID), d.logger.Field("error", err.Error()))
}

// Event logs general events
func (d *CustomDebugger) Event(e *debug.Event) {
	// Log the event using your logger
	d.logger.Info("Event", d.logger.Field("type", e.Type), d.logger.Field("request_id", e.RequestID), d.logger.Field("collector_id", e.CollectorID))
}
