// Package config provides configuration management for the application.
package config

import "time"

// Default configuration values
const (
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default logging format
	DefaultLogFormat = "json"

	// DefaultLogOutput is the default logging output
	DefaultLogOutput = "stdout"

	// DefaultLogMaxSize is the default maximum size of log files in megabytes
	DefaultLogMaxSize = 100

	// DefaultLogMaxBackups is the default maximum number of log file backups
	DefaultLogMaxBackups = 3

	// DefaultLogMaxAge is the default maximum age of log files in days
	DefaultLogMaxAge = 28

	// DefaultLogCompress is the default value for log file compression
	DefaultLogCompress = true

	// DefaultStorageType is the default storage type
	DefaultStorageType = "elasticsearch"

	// DefaultHTTPPort is the default HTTP port
	DefaultHTTPPort = 8080

	// DefaultHTTPHost is the default HTTP host
	DefaultHTTPHost = "localhost"

	// DefaultHTTPTimeout is the default HTTP timeout
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPReadTimeout is the default HTTP read timeout
	DefaultHTTPReadTimeout = 15 * time.Second

	// DefaultHTTPWriteTimeout is the default HTTP write timeout
	DefaultHTTPWriteTimeout = 15 * time.Second

	// DefaultHTTPIdleTimeout is the default HTTP idle timeout
	DefaultHTTPIdleTimeout = 60 * time.Second
)
