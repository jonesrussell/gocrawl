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

	// DefaultRandomDelayFactor is used to calculate random delay for rate limiting
	DefaultRandomDelayFactor = 2

	// DefaultParallelism is the default number of parallel requests
	DefaultParallelism = 2

	// DefaultStartTimeout is the default timeout for starting the crawler
	DefaultStartTimeout = 30 * time.Second

	// DefaultStopTimeout is the default timeout for stopping the crawler
	DefaultStopTimeout = 30 * time.Second

	// DefaultPollInterval is the default interval for polling crawler status
	DefaultPollInterval = 100 * time.Millisecond

	// DefaultMaxRetries is the default number of retries for failed requests
	DefaultMaxRetries = 3

	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 2

	// DefaultRateLimit is the default rate limit for requests
	DefaultRateLimit = 2 * time.Second

	// DefaultRandomDelay is the default random delay between requests
	DefaultRandomDelay = 5 * time.Second

	// DefaultBufferSize is the default size for channel buffers
	DefaultBufferSize = 100

	// DefaultMaxConcurrency is the default maximum number of concurrent requests
	DefaultMaxConcurrency = 2

	// DefaultTestSleepDuration is the default sleep duration for tests
	DefaultTestSleepDuration = 100 * time.Millisecond

	// DefaultZapFieldsCapacity is the default capacity for zap fields slice.
	DefaultZapFieldsCapacity = 2

	// CollectorStartTimeout is the timeout for collector initialization
	CollectorStartTimeout = 5 * time.Second
)
