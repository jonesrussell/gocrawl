package crawler

import (
	"github.com/jonesrussell/gocrawl/internal/constants"
)

// Re-export crawler constants from constants package for backward compatibility
const (
	DefaultArticleChannelBufferSize = constants.DefaultArticleChannelBufferSize
	CrawlerStartTimeout             = constants.CrawlerStartTimeout
	DefaultStopTimeout              = constants.DefaultStopTimeout
	CrawlerPollInterval             = constants.CrawlerPollInterval
	CrawlerCollectorStartTimeout    = constants.CrawlerCollectorStartTimeout
	DefaultProcessorsCapacity       = constants.DefaultProcessorsCapacity
)
