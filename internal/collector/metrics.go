// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import "time"

// Metrics contains statistics about the crawling process.
type Metrics struct {
	// PagesVisited is the total number of pages visited.
	PagesVisited int64
	// ArticlesFound is the total number of articles found.
	ArticlesFound int64
	// ContentFound is the total number of content pages found.
	ContentFound int64
	// Errors is the total number of errors encountered.
	Errors int64
	// StartTime is when the crawler started.
	StartTime int64
	// EndTime is when the crawler finished.
	EndTime int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now().Unix(),
	}
}
