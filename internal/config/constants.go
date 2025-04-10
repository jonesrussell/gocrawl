// Package config provides configuration management for the GoCrawl application.
package config

import (
	"time"
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

	// DefaultServerAddress is the default HTTP server address
	DefaultServerAddress = ":8080"

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

	// defaultRetryMaxWait is the default maximum wait time between retries
	defaultRetryMaxWait = 30 * time.Second

	// defaultRetryInitialWait is the default initial wait time between retries
	defaultRetryInitialWait = 1 * time.Second

	// defaultMaxRetries is the default number of retries for failed requests
	defaultMaxRetries = 3

	// DefaultServerPort is the default server port
	DefaultServerPort = 8080

	// Constants for default configuration values
	defaultMaxAge             = 86400 // 24 hours in seconds
	defaultRateLimitPerMinute = 60

	// Default rate limits
	defaultCrawlerRateLimit = "1s"
	defaultRandomDelay      = 500 * time.Millisecond

	// Default Elasticsearch settings
	defaultESAddress = "http://localhost:9200"
	defaultESIndex   = "gocrawl"

	// Default app settings
	defaultAppName    = "gocrawl"
	defaultAppVersion = "1.0.0"
	defaultAppEnv     = "development"

	// Default values for various configurations
	DefaultMaxRetries          = 3
	DefaultBulkSize            = 1000
	DefaultFlushInterval       = 30 * time.Second
	DefaultPriority            = 5
	DefaultMaxPriority         = 10
	DefaultTimeout             = 10 * time.Second
	DefaultMaxHeaderBytes      = 1 << 20            // 1 MB
	DefaultStorageMaxSize      = 1024 * 1024 * 1024 // 1 GB
	DefaultStorageMaxItems     = 10000
	DefaultMaxIdleConns        = 100
	DefaultIdleConnTimeout     = 90 * time.Second
	DefaultTLSHandshakeTimeout = 10 * time.Second
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

// Default Elasticsearch settings
const (
	DefaultElasticsearchHost     = "http://localhost:9200"
	DefaultElasticsearchIndex    = "gocrawl"
	DefaultElasticsearchUsername = "elastic"
	DefaultElasticsearchPassword = "changeme"
)
