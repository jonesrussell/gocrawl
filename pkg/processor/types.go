// Package processor provides content processing functionality for the application.
package processor

import (
	"github.com/jonesrussell/gocrawl/internal/common"
)

// HTMLProcessor processes HTML content.
type HTMLProcessor struct {
	// Selectors are the CSS selectors to use.
	selectors map[string]string
	// Logger is the logger for the processor.
	logger common.Logger
	// MetricsCollector is the metrics collector for the processor.
	metrics MetricsCollector
}

// NewHTMLProcessor creates a new HTML processor.
func NewHTMLProcessor(selectors map[string]string, logger common.Logger, metrics MetricsCollector) *HTMLProcessor {
	return &HTMLProcessor{
		selectors: selectors,
		logger:    logger,
		metrics:   metrics,
	}
}
