// Package crawler provides the core crawling functionality for the application.
package crawler

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the crawler interface
type Interface interface {
	// Start starts the crawler for the given source
	Start(ctx context.Context, sourceName string) error
	// Stop stops the crawler
	Stop(ctx context.Context) error
	// Subscribe subscribes to crawler events
	Subscribe(handler events.Handler)
	// SetRateLimit sets the rate limit for the crawler
	SetRateLimit(duration time.Duration) error
	// SetMaxDepth sets the maximum depth for the crawler
	SetMaxDepth(depth int)
	// SetCollector sets the collector for the crawler
	SetCollector(collector *colly.Collector)
	// GetIndexManager returns the index manager
	GetIndexManager() storagetypes.IndexManager
	// Wait waits for the crawler to complete
	Wait()
	// GetMetrics returns the crawler metrics
	GetMetrics() *common.Metrics
	// SetTestServerURL sets the test server URL
	SetTestServerURL(url string)
}
