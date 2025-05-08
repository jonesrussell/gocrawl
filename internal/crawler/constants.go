package crawler

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
)

const (
	// DefaultArticleChannelBufferSize is the default buffer size for the article channel.
	DefaultArticleChannelBufferSize = common.DefaultBufferSize

	// CrawlerStartTimeout is the default timeout for starting the crawler
	CrawlerStartTimeout = 30 * time.Second

	// DefaultStopTimeout is the default timeout for stopping the crawler
	DefaultStopTimeout = 30 * time.Second

	// CrawlerPollInterval is the default interval for polling crawler status
	CrawlerPollInterval = 100 * time.Millisecond

	// CrawlerCollectorStartTimeout is the timeout for collector initialization
	CrawlerCollectorStartTimeout = 5 * time.Second

	// DefaultProcessorsCapacity is the default capacity for processor slices.
	DefaultProcessorsCapacity = 2
)
