package crawler

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
)

const (
	// DefaultArticleChannelBufferSize is the default buffer size for the article channel.
	DefaultArticleChannelBufferSize = common.DefaultBufferSize

	// DefaultHTTPPort is the default HTTP port.
	DefaultHTTPPort = 8080

	// DefaultReadTimeout is the default read timeout.
	DefaultReadTimeout = 15 * time.Second

	// DefaultWriteTimeout is the default write timeout.
	DefaultWriteTimeout = 15 * time.Second

	// DefaultIdleTimeout is the default idle timeout.
	DefaultIdleTimeout = 90 * time.Second

	// DefaultMaxHeaderBytes is the default maximum header bytes.
	DefaultMaxHeaderBytes = 1 << 20

	// DefaultRequestTimeout is the default request timeout.
	DefaultRequestTimeout = common.DefaultOperationTimeout

	// DefaultDelay is the default delay between requests.
	DefaultDelay = common.DefaultRateLimit

	// DefaultDefaultPriority is the default priority for jobs.
	DefaultDefaultPriority = 5

	// DefaultMaxPriority is the default maximum priority.
	DefaultMaxPriority = 10

	// DefaultBulkSize is the default bulk size for operations.
	DefaultBulkSize = 1000

	// DefaultFlushInterval is the default flush interval.
	DefaultFlushInterval = 30 * time.Second
)
