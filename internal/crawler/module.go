// Package crawler provides core crawling functionality.
package crawler

import (
	"context"

	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"go.uber.org/fx"
)

// Interface defines the crawler's capabilities.
type Interface interface {
	// Start begins crawling from the given base URL.
	Start(ctx context.Context, baseURL string) error
	// Stop gracefully stops the crawler.
	Stop(ctx context.Context) error
	// Subscribe adds a content handler to receive discovered content.
	Subscribe(handler events.Handler)
	// SetRateLimit sets the crawler's rate limit.
	SetRateLimit(duration string) error
	// SetMaxDepth sets the maximum crawl depth.
	SetMaxDepth(depth int)
}

// Module provides the crawler's dependencies.
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		provideEventBus,
		provideCrawler,
	),
)

// Params defines the crawler's required dependencies.
type Params struct {
	fx.In

	Logger   common.Logger
	Debugger debug.Debugger `optional:"true"`
}

// Result contains the crawler's provided components.
type Result struct {
	fx.Out

	Crawler Interface
}

// provideEventBus creates a new event bus instance.
func provideEventBus() *events.Bus {
	return events.NewBus()
}

// provideCollyDebugger creates a new debugger instance.
func provideCollyDebugger(logger common.Logger) debug.Debugger {
	return &debug.LogDebugger{
		Output: newDebugLogger(logger),
	}
}

// provideCrawler creates a new crawler instance.
func provideCrawler(p Params, bus *events.Bus) (Result, error) {
	c := &Crawler{
		Logger:   p.Logger,
		Debugger: p.Debugger,
		bus:      bus,
	}
	return Result{Crawler: c}, nil
}
