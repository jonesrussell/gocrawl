// Package config provides configuration management for the GoCrawl application.
package config

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
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

// Default configuration values
const (
	// DefaultRateLimit is the default delay between requests
	DefaultRateLimit = 2 * time.Second
	// DefaultMaxDepth is the default maximum crawl depth
	DefaultMaxDepth = 3
	// DefaultParallelism is the default number of concurrent crawlers
	DefaultParallelism = 2

	// DefaultReadTimeout is the default HTTP server read timeout
	DefaultReadTimeout = 15 * time.Second

	// DefaultWriteTimeout is the default HTTP server write timeout
	DefaultWriteTimeout = 15 * time.Second

	// DefaultIdleTimeout is the default HTTP server idle timeout
	DefaultIdleTimeout = 120 * time.Second

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"

	// DefaultEnvironment is the default application environment
	DefaultEnvironment = "development"

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

	// DefaultStorageType is the default storage backend type
	DefaultStorageType = "elasticsearch"

	// DefaultStorageBatchSize is the default number of items to batch for storage operations
	DefaultStorageBatchSize = 100

	// DefaultElasticsearchRetries is the default number of retries for Elasticsearch operations
	DefaultElasticsearchRetries = 3

	// DefaultBulkSize is the default number of documents to bulk index
	DefaultBulkSize = elasticsearch.DefaultBulkSize

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

	// DefaultRetryMaxWait is the default maximum wait time between retries
	DefaultRetryMaxWait = 30 * time.Second

	// DefaultRetryInitialWait is the default initial wait time between retries
	DefaultRetryInitialWait = 1 * time.Second

	// DefaultMaxRetries is the default number of retries for failed requests
	DefaultMaxRetries = 3

	// DefaultServerPort is the default server port
	DefaultServerPort = 8080

	// DefaultMaxAge is the default maximum age in seconds (24 hours)
	DefaultMaxAge = 86400

	// DefaultRateLimitPerMinute is the default rate limit per minute
	DefaultRateLimitPerMinute = 60

	// DefaultCrawlerRateLimit is the default crawler rate limit
	DefaultCrawlerRateLimit = "1s"

	// DefaultRandomDelay is the default random delay between requests
	DefaultRandomDelay = 500 * time.Millisecond

	// DefaultESAddress is the default Elasticsearch address
	DefaultESAddress = "http://localhost:9200"

	// DefaultESIndex is the default Elasticsearch index name
	DefaultESIndex = "gocrawl"

	// DefaultAppName is the default application name
	DefaultAppName = "gocrawl"

	// DefaultAppVersion is the default application version
	DefaultAppVersion = "1.0.0"

	// DefaultAppEnv is the default application environment
	DefaultAppEnv = "development"

	// DefaultFlushInterval is the default flush interval for Elasticsearch operations
	DefaultFlushInterval = 30 * time.Second

	// DefaultPriority is the default priority for items
	DefaultPriority = 5

	// DefaultMaxPriority is the default maximum priority
	DefaultMaxPriority = 10

	// DefaultTimeout is the default timeout for operations
	DefaultTimeout = 10 * time.Second

	// DefaultMaxHeaderBytes is the default maximum header bytes (1 MB)
	DefaultMaxHeaderBytes = 1 << 20

	// DefaultStorageMaxSize is the default maximum storage size (1 GB)
	DefaultStorageMaxSize = 1024 * 1024 * 1024

	// DefaultStorageMaxItems is the default maximum number of items to store
	DefaultStorageMaxItems = 10000

	// DefaultMaxIdleConns is the default maximum number of idle (keep-alive) connections
	DefaultMaxIdleConns = 100

	// DefaultMaxIdleConnsPerHost is the default maximum number of idle (keep-alive) connections per host
	DefaultMaxIdleConnsPerHost = 100

	// DefaultIdleConnTimeout is the default maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing itself
	DefaultIdleConnTimeout = 90 * time.Second

	// DefaultTLSHandshakeTimeout is the default maximum amount of time waiting to wait for a TLS handshake
	DefaultTLSHandshakeTimeout = 10 * time.Second

	// DefaultMaxBodySize is the default maximum body size (10MB)
	DefaultMaxBodySize = 10 * 1024 * 1024

	// DefaultMaxRateLimitCount is the default maximum rate limit count
	DefaultMaxRateLimitCount = 100

	// Source-specific defaults
	DefaultSourceMaxDepth  = 2
	DefaultSourceRateLimit = 5 * time.Second
)

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

// Rule actions
const (
	// ActionAllow indicates that a URL pattern should be allowed
	ActionAllow = "allow"
	// ActionDisallow indicates that a URL pattern should be disallowed
	ActionDisallow = "disallow"
)

// ValidRuleActions contains all valid rule actions
var ValidRuleActions = map[string]bool{
	ActionAllow:    true,
	ActionDisallow: true,
}

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
