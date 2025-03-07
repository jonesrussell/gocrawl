package collector

import (
	"github.com/gocolly/colly/v2"
)

// CollectorHandlers manages all event handlers for the collector
type CollectorHandlers struct {
	config *CollectorConfig
	done   chan struct{}
}

// NewCollectorHandlers creates a new collector handlers instance
func NewCollectorHandlers(config *CollectorConfig, done chan struct{}) *CollectorHandlers {
	return &CollectorHandlers{
		config: config,
		done:   done,
	}
}

// ConfigureHandlers sets up all event handlers for the collector
func (h *CollectorHandlers) ConfigureHandlers(c *colly.Collector) {
	// Add completion handler to ensure proper completion signaling
	c.OnScraped(func(r *colly.Response) {
		// Check if this is the last request
		if r.Request.URL.String() == h.config.BaseURL {
			h.config.Logger.Debug("Base URL scraped, crawl complete")
			// Signal completion by closing the done channel
			close(h.done)
		}
	})

	// Add error handler to ensure we know about any failures
	c.OnError(func(r *colly.Response, err error) {
		h.config.Logger.Error("Request failed", "url", r.Request.URL, "error", err)
	})

	// Add request handler to track progress
	c.OnRequest(func(r *colly.Request) {
		h.config.Logger.Debug("Starting request", "url", r.URL)
	})

	// Add response handler to track completion
	c.OnResponse(func(r *colly.Response) {
		h.config.Logger.Debug("Received response", "url", r.Request.URL, "status", r.StatusCode)
	})
}
