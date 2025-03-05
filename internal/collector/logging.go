package collector

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ConfigureLogging sets up logging for the collector
func ConfigureLogging(c *colly.Collector, log logger.Interface) {
	c.OnRequest(func(r *colly.Request) {
		log.Debug("Requesting URL", "url", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Debug("Received response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Error("Error occurred",
			"url", r.Request.URL.String(),
			"error", err.Error(),
		)
	})
}
