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
	// collector is the Colly collector instance
	collector *colly.Collector
}

// NewHandlers creates a new collector handlers instance.
// It initializes the handlers with the provided configuration and
// completion channel.
//
// Parameters:
//   - config: The collector configuration
//   - done: Channel to signal when crawling is complete
//   - collector: The Colly collector instance to configure
//
// Returns:
//   - *Handlers: The initialized handlers instance
func NewHandlers(config *Config, done chan struct{}, collector *colly.Collector) *Handlers {
	return &Handlers{
		config:    config,
		done:      done,
		collector: collector,
	}
}

// ConfigureHandlers sets up all event handlers for the collector.
// It:
// - Sets up handlers for scraping completion
// - Configures error handling
// - Sets up request and response tracking
// - Manages crawl completion signaling
func (h *Handlers) ConfigureHandlers() {
	// Handle scraping completion
	h.collector.OnScraped(func(r *colly.Response) {
		h.config.Logger.Debug("Scraped URL", "url", r.Request.URL.String())
	})

	// Handle errors
	h.collector.OnError(func(r *colly.Response, err error) {
		h.config.Logger.Error("Error scraping URL",
			"url", r.Request.URL.String(),
			"error", err,
		)
	})

	// Handle requests
	h.collector.OnRequest(func(r *colly.Request) {
		h.config.Logger.Debug("Starting request", "url", r.URL.String())
	})

	// Handle responses
	h.collector.OnResponse(func(r *colly.Response) {
		h.config.Logger.Debug("Received response", "url", r.Request.URL.String())
	})

	// Process content with article processor
	if h.config.ArticleProcessor != nil {
		h.collector.OnHTML("*", func(e *colly.HTMLElement) {
			if err := h.config.ArticleProcessor.Process(e); err != nil {
				h.config.Logger.Error("Failed to process article content",
					"url", e.Request.URL.String(),
					"error", err)
			}
		})
	}

	// Process content with content processor
	if h.config.ContentProcessor != nil {
		h.collector.OnHTML("*", func(e *colly.HTMLElement) {
			if err := h.config.ContentProcessor.Process(e); err != nil {
				h.config.Logger.Error("Failed to process content",
					"url", e.Request.URL.String(),
					"error", err)
			}
		})
	}
}
