// Package processor provides content processing functionality for the application.
package processor

import (
	"github.com/jonesrussell/gocrawl/pkg/logger"
)

// HTMLProcessor processes HTML content.
type HTMLProcessor struct {
	// Selectors are the CSS selectors to use.
	Selectors map[string]string
	// Logger is the logger for the processor.
	Logger logger.Interface
	// MetricsCollector is the metrics collector for the processor.
	MetricsCollector MetricsCollector
}

// NewHTMLProcessor creates a new HTML processor.
func NewHTMLProcessor(selectors map[string]string, logger logger.Interface, metrics MetricsCollector) *HTMLProcessor {
	return &HTMLProcessor{
		Selectors:        selectors,
		Logger:           logger,
		MetricsCollector: metrics,
	}
}
