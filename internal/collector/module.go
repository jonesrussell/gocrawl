// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector module for dependency injection.
// It provides:
// - Collector configuration
// - Collector instance
// - Content processors
// - Metrics
var Module = fx.Module("collector",
	fx.Provide(
		NewConfig,
		New,
		NewMetrics,
	),
)
