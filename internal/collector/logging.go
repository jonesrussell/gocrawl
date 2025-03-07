// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ConfigureLogging sets up logging for the collector.
// It configures event handlers for various collector events to provide
// detailed logging information about the crawling process.
//
// The function sets up handlers for:
// - Request events: Logs when a URL is being requested
// - Response events: Logs when a response is received, including status code
// - Error events: Logs any errors that occur during crawling
//
// Parameters:
//   - c: The Colly collector instance to configure
//   - log: The logger interface to use for logging events
func ConfigureLogging(c *colly.Collector, log logger.Interface) {
	// Log when a request is about to be made
	c.OnRequest(func(r *colly.Request) {
		log.Debug("Requesting URL", "url", r.URL.String())
	})

	// Log when a response is received
	c.OnResponse(func(r *colly.Response) {
		log.Debug("Received response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
		)
	})

	// Log any errors that occur during crawling
	c.OnError(func(r *colly.Response, err error) {
		log.Error("Error occurred",
			"url", r.Request.URL.String(),
			"error", err.Error(),
		)
	})
}
