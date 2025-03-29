// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
//
// Key features:
// - Configurable crawling depth and parallelism
// - Rate limiting and politeness delays
// - Content extraction and processing
// - Metrics collection and monitoring
// - Error handling and retries
//
// The package integrates with the article and content packages to process
// different types of web content encountered during crawling.
package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector module for dependency injection.
// It provides:
// - Collector configuration with environment variable support
// - Collector instance with configured settings
// - Metrics collection and monitoring
//
// The module uses fx.Provide to wire up dependencies and ensure proper
// initialization of the collector components. Configuration is loaded from
// environment variables with sensible defaults.
var Module = fx.Module("collector",
	fx.Provide(
		NewConfig,
		New,
		NewMetrics,
	),
)
