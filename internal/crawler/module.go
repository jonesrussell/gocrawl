// Package crawler provides the core crawling functionality for GoCrawl.
// It manages the crawling process, including URL processing, rate limiting,
// and content extraction.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
)

// Interface defines the crawler's capabilities.
type Interface interface {
	// Start begins crawling from the given base URL.
	Start(ctx context.Context, sourceName string) error
	// Stop gracefully stops the crawler.
	Stop(ctx context.Context) error
	// Subscribe adds a content handler to receive discovered content.
	Subscribe(handler events.Handler)
	// SetRateLimit sets the crawler's rate limit.
	SetRateLimit(duration time.Duration) error
	// SetMaxDepth sets the maximum crawl depth.
	SetMaxDepth(depth int)
	// SetCollector sets the collector for the crawler.
	SetCollector(collector *colly.Collector)
	// GetIndexManager returns the index manager interface.
	GetIndexManager() api.IndexManager
	// Wait blocks until the crawler has finished processing all queued requests.
	Wait()
	// GetMetrics returns the current crawler metrics.
	GetMetrics() *common.Metrics
}

// Result defines the crawler module's output.
type Result struct {
	fx.Out
	Crawler Interface
}

// Module provides the crawler module and its dependencies.
var Module = fx.Module("crawler",
	fx.Provide(
		// Provide the crawler implementation
		fx.Annotate(
			ProvideCrawler,
			fx.ParamTags(
				``,
				``,
				``,
				``,
				`name:"startupArticleProcessor"`,
				`name:"startupContentProcessor"`,
				``,
			),
		),
	),
)

// ProvideCrawler creates a new crawler instance with all dependencies.
func ProvideCrawler(
	logger common.Logger,
	debugger debug.Debugger,
	indexManager api.IndexManager,
	sources *sources.Sources,
	articleProcessor common.Processor,
	contentProcessor common.Processor,
	bus *events.Bus,
) Result {
	return Result{
		Crawler: &Crawler{
			Logger:           logger,
			Debugger:         debugger,
			indexManager:     indexManager,
			sources:          sources,
			articleProcessor: articleProcessor,
			contentProcessor: contentProcessor,
			bus:              bus,
		},
	}
}
