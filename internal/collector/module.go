// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/api"
	"go.uber.org/fx"
)

// Module provides the collector module and its dependencies.
// It uses fx.Module to define the collector package as a dependency injection module,
// providing the New function as a constructor for creating collector instances.
//
// The module is used to integrate the collector package into the main application
// using the fx dependency injection framework.
func Module() fx.Option {
	return fx.Module("collector",
		fx.Provide(
			fx.Annotate(
				New,
				fx.As(new(Interface)),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewConfig,
				fx.As(new(*Config)),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewHandlers,
				fx.As(new(*Handlers)),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewSetup,
				fx.As(new(*Setup)),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewMetrics,
				fx.As(new(*Metrics)),
			),
		),
	)
}

// Interface defines the collector's capabilities.
type Interface interface {
	// Visit starts crawling from the given URL.
	Visit(url string) error
	// Wait blocks until the collector has finished processing all queued requests.
	Wait()
	// SetRateLimit sets the collector's rate limit.
	SetRateLimit(duration time.Duration) error
	// SetMaxDepth sets the maximum crawl depth.
	SetMaxDepth(depth int)
	// SetCollector sets the collector for the crawler.
	SetCollector(collector *colly.Collector)
	// GetIndexManager returns the index manager interface.
	GetIndexManager() api.IndexManager
	// GetMetrics returns the current crawler metrics.
	GetMetrics() *Metrics
}
