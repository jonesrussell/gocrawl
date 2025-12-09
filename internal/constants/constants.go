// Package constants provides all shared constants used across the GoCrawl application.
// Constants are organized by domain (HTTP, Crawler, Storage, Logger, General).
package constants

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
)

// HTTP/Server Constants
const (
	// DefaultServerAddress is the default HTTP server address
	DefaultServerAddress = ":8080"

	// DefaultServerReadTimeout is the default HTTP server read timeout
	DefaultServerReadTimeout = 30 * time.Second

	// DefaultServerWriteTimeout is the default HTTP server write timeout
	DefaultServerWriteTimeout = 30 * time.Second

	// DefaultServerIdleTimeout is the default HTTP server idle timeout
	DefaultServerIdleTimeout = 60 * time.Second

	// DefaultReadTimeout is the default HTTP server read timeout
	DefaultReadTimeout = 15 * time.Second

	// DefaultWriteTimeout is the default HTTP server write timeout
	DefaultWriteTimeout = 15 * time.Second

	// DefaultIdleTimeout is the default HTTP server idle timeout
	DefaultIdleTimeout = 120 * time.Second

	// DefaultHTTPPort is the default HTTP server port
	DefaultHTTPPort = 8080

	// DefaultHTTPHost is the default HTTP server host
	DefaultHTTPHost = "localhost"

	// DefaultHTTPTimeout is the default timeout for HTTP requests
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPReadTimeout is the default read timeout for HTTP requests
	DefaultHTTPReadTimeout = 15 * time.Second

	// DefaultHTTPWriteTimeout is the default write timeout for HTTP responses
	DefaultHTTPWriteTimeout = 15 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout for HTTP connections
	DefaultHTTPIdleTimeout = 90 * time.Second

	// DefaultMaxHeaderBytes is the default maximum header bytes (1 MB)
	DefaultMaxHeaderBytes = 1 << 20

	// DefaultMaxBodySize is the default maximum body size (10MB)
	DefaultMaxBodySize = 10 * 1024 * 1024

	// Transport constants
	DefaultMaxIdleConns          = 100
	DefaultMaxIdleConnsPerHost   = 10
	DefaultIdleConnTimeout       = 90 * time.Second
	DefaultTLSHandshakeTimeout   = 10 * time.Second
	DefaultResponseHeaderTimeout = 30 * time.Second
	DefaultExpectContinueTimeout = 1 * time.Second
)

// Crawler Constants
const (
	// DefaultRateLimit is the default delay between requests
	DefaultRateLimit = 2 * time.Second

	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 3

	// DefaultParallelism is the default number of concurrent crawlers
	DefaultParallelism = 2

	// DefaultBufferSize is the default size for channel buffers
	DefaultBufferSize = 100

	// DefaultMaxConcurrency is the default maximum number of concurrent requests
	DefaultMaxConcurrency = 2

	// DefaultArticleChannelBufferSize is the default buffer size for the article channel.
	DefaultArticleChannelBufferSize = DefaultBufferSize

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

	// DefaultCrawlerRateLimit is the default crawler rate limit
	DefaultCrawlerRateLimit = "1s"

	// DefaultRandomDelay is the default random delay between requests
	DefaultRandomDelay = 500 * time.Millisecond

	// DefaultSourceMaxDepth is the default maximum depth for sources
	DefaultSourceMaxDepth = 2

	// DefaultSourceRateLimit is the default rate limit for sources
	DefaultSourceRateLimit = 5 * time.Second

	// DefaultMaxAge is the default maximum age in seconds (24 hours)
	DefaultMaxAge = 86400

	// DefaultRateLimitPerMinute is the default rate limit per minute
	DefaultRateLimitPerMinute = 60
)

// Storage Constants
const (
	// DefaultStorageType is the default storage backend type
	DefaultStorageType = "elasticsearch"

	// DefaultStorageBatchSize is the default number of items to batch for storage operations
	DefaultStorageBatchSize = 100

	// DefaultElasticsearchRetries is the default number of retries for Elasticsearch operations
	DefaultElasticsearchRetries = 3

	// DefaultBulkSize is the default number of documents to bulk index
	DefaultBulkSize = elasticsearch.DefaultBulkSize

	// DefaultESAddress is the default Elasticsearch address
	DefaultESAddress = "http://localhost:9200"

	// DefaultESIndex is the default Elasticsearch index name
	DefaultESIndex = "gocrawl"

	// DefaultStorageMaxSize is the default maximum storage size (1 GB)
	DefaultStorageMaxSize = 1024 * 1024 * 1024

	// DefaultStorageMaxItems is the default maximum number of items to store
	DefaultStorageMaxItems = 10000

	// DefaultFlushInterval is the default flush interval for Elasticsearch operations
	DefaultFlushInterval = 30 * time.Second
)

// Logger Constants
const (
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default logging format
	DefaultLogFormat = "json"

	// DefaultLogOutput is the default logging output
	DefaultLogOutput = "stdout"

	// DefaultLogMaxSize is the default maximum size in megabytes of the log file before it gets rotated
	DefaultLogMaxSize = 100

	// DefaultLogMaxBackups is the default maximum number of old log files to retain
	DefaultLogMaxBackups = 3

	// DefaultLogMaxAge is the default maximum number of days to retain old log files
	DefaultLogMaxAge = 30

	// DefaultLogCompress determines if the rotated log files should be compressed
	DefaultLogCompress = true

	// DefaultMaxLogSize is the default maximum size of a log file in MB
	DefaultMaxLogSize = 100

	// DefaultMaxLogBackups is the default number of log file backups to keep
	DefaultMaxLogBackups = 3

	// DefaultMaxLogAge is the default maximum age of a log file in days
	DefaultMaxLogAge = 30
)

// General/Common Constants
const (
	// DefaultOperationTimeout is the default timeout for general operations.
	// This duration is used for common operations like API calls,
	// data processing tasks, or crawler shutdown that should complete
	// in a reasonable time.
	DefaultOperationTimeout = 30 * time.Second

	// DefaultCrawlerTimeout is the maximum time to wait for the crawler to complete.
	DefaultCrawlerTimeout = 30 * time.Minute

	// DefaultShutdownTimeout is the maximum time to wait for graceful shutdown.
	DefaultShutdownTimeout = 30 * time.Second

	// DefaultArticleIndex is the default index name for articles
	DefaultArticleIndex = "articles"

	// DefaultPageIndex is the default index name for pages
	DefaultPageIndex = "pages"

	// DefaultContentIndex is the default index name for general content
	DefaultContentIndex = "content"

	// DefaultIndicesCapacity is the initial capacity for index slices.
	// Set to 2 to accommodate both content and article indices for a source.
	DefaultIndicesCapacity = 2

	// DefaultTestSleepDuration is the default sleep duration for tests
	DefaultTestSleepDuration = 100 * time.Millisecond

	// DefaultMaxRetries is the default number of retries for failed requests
	DefaultMaxRetries = 3

	// DefaultRetryMaxWait is the default maximum wait time between retries
	DefaultRetryMaxWait = 30 * time.Second

	// DefaultRetryInitialWait is the default initial wait time between retries
	DefaultRetryInitialWait = 1 * time.Second

	// DefaultTimeout is the default timeout for operations
	DefaultTimeout = 10 * time.Second

	// DefaultMaxRateLimitCount is the default maximum rate limit count
	DefaultMaxRateLimitCount = 100

	// DefaultPriority is the default priority for items
	DefaultPriority = 5

	// DefaultMaxPriority is the default maximum priority
	DefaultMaxPriority = 10

	// DefaultEnvironment is the default application environment
	DefaultEnvironment = "development"

	// DefaultAppName is the default application name
	DefaultAppName = "gocrawl"

	// DefaultAppVersion is the default application version
	DefaultAppVersion = "1.0.0"

	// DefaultAppEnv is the default application environment
	DefaultAppEnv = "development"

	// MinArticleBodyLength is the minimum number of characters required
	// for a page to be detected as an article based on body content
	MinArticleBodyLength = 200
)

// Environment types
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Storage types
const (
	StorageTypeElasticsearch = "elasticsearch"
	StorageTypeFile          = "file"
	StorageTypeMemory        = "memory"
)

// Rule actions
const (
	// ActionAllow indicates that a URL pattern should be allowed
	ActionAllow = "allow"
	// ActionDisallow indicates that a URL pattern should be disallowed
	ActionDisallow = "disallow"
)

// ValidLogLevels defines the valid logging levels
var ValidLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// ValidEnvironments defines the valid environment types
var ValidEnvironments = map[string]bool{
	"development": true,
	"staging":     true,
	"production":  true,
	"test":        true,
}

// ValidHTTPMethods defines the valid HTTP methods
var ValidHTTPMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

// ValidHTTPHeaders defines the valid HTTP headers
var ValidHTTPHeaders = map[string]bool{
	"Accept":            true,
	"Accept-Charset":    true,
	"Accept-Encoding":   true,
	"Accept-Language":   true,
	"Authorization":     true,
	"Cache-Control":     true,
	"Connection":        true,
	"Content-Length":    true,
	"Content-Type":      true,
	"Cookie":            true,
	"Host":              true,
	"Origin":            true,
	"Referer":           true,
	"User-Agent":        true,
	"X-Forwarded-For":   true,
	"X-Forwarded-Proto": true,
	"X-Real-IP":         true,
	"X-Request-ID":      true,
}

// ValidRuleActions contains all valid rule actions
var ValidRuleActions = map[string]bool{
	ActionAllow:    true,
	ActionDisallow: true,
}
