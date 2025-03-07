// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"github.com/gocolly/colly/v2"
)

// Handlers manages all event handlers for the collector.
// It coordinates the handling of various collector events and manages
// the completion signaling through a done channel.
type Handlers struct {
	// config contains the collector configuration
	config *Config
	// done is a channel used to signal when crawling is complete
	done chan struct{}
}

// NewHandlers creates a new collector handlers instance.
// It initializes the handlers with the provided configuration and
// completion channel.
//
// Parameters:
//   - config: The collector configuration
//   - done: Channel to signal when crawling is complete
//
// Returns:
//   - *Handlers: The initialized handlers instance
func NewHandlers(config *Config, done chan struct{}) *Handlers {
	return &Handlers{
		config: config,
		done:   done,
	}
}

// ConfigureHandlers sets up all event handlers for the collector.
// It configures handlers for various collector events including:
// - Scraping completion
// - Error handling
// - Request tracking
// - Response tracking
//
// Parameters:
//   - c: The Colly collector instance to configure
func (h *Handlers) ConfigureHandlers(c *colly.Collector) {
	// Add completion handler to ensure proper completion signaling
	c.OnScraped(func(r *colly.Response) {
		// Check if this is the last request (base URL)
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
