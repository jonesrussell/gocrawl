// Package common provides shared functionality and interfaces used across
// the GoCrawl application.
package common

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// CollyDebugger implements the colly.Debugger interface to provide
// debug logging for Colly operations using our logger.
type CollyDebugger struct {
	logger logger.Interface
}

// NewCollyDebugger creates a new Colly debugger that uses the provided logger.
// It wraps our logger interface to provide debug logging for Colly operations.
//
// Parameters:
//   - logger: The logger to use for debug messages
//
// Returns:
//   - *CollyDebugger: A new debugger instance
func NewCollyDebugger(logger logger.Interface) *CollyDebugger {
	return &CollyDebugger{
		logger: logger,
	}
}

// Init implements the colly.Debugger interface.
// It is called when the debugger is initialized.
func (d *CollyDebugger) Init() error {
	return nil
}

// Event implements the colly.Debugger interface.
// It logs debug events from Colly using our common logger.
func (d *CollyDebugger) Event(e interface{}) {
	d.logger.Debug("Colly event",
		"event", e,
	)
}

// Error implements the colly.Debugger interface.
// It logs debug errors from Colly using our common logger.
func (d *CollyDebugger) Error(e interface{}) {
	d.logger.Error("Colly debug error",
		"error", e,
	)
}

// Request implements the colly.Debugger interface.
// It logs debug requests from Colly using our common logger.
func (d *CollyDebugger) Request(req *colly.Request) {
	d.logger.Debug("Colly request",
		"url", req.URL.String(),
		"method", req.Method,
		"headers", req.Headers,
	)
}

// Response implements the colly.Debugger interface.
// It logs debug responses from Colly using our common logger.
func (d *CollyDebugger) Response(resp *colly.Response) {
	d.logger.Debug("Colly response",
		"url", resp.Request.URL.String(),
		"status", resp.StatusCode,
		"headers", resp.Headers,
	)
}
