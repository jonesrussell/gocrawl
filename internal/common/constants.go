// Package common provides shared functionality, constants, and utilities
// used across the GoCrawl application. This package serves as a central
// location for common types, interfaces, and helper functions.
package common

import "time"

// Timeout constants define the default durations for various operations
// throughout the application. These values can be overridden through
// configuration if needed.
const (
	// DefaultOperationTimeout is the default timeout for general operations.
	// This duration is used for common operations like API calls,
	// data processing tasks, or crawler shutdown that should complete
	// in a reasonable time.
	DefaultOperationTimeout = 30 * time.Second
)

const (
	// DefaultTestSleepDuration is the default sleep duration for tests
	DefaultTestSleepDuration = 100 * time.Millisecond
	// DefaultMaxRetries is the default number of retries for failed requests
	DefaultMaxRetries = 3
	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 2
	// DefaultRateLimit is the default rate limit for requests
	DefaultRateLimit = 2 * time.Second
	// DefaultBufferSize is the default size for channel buffers
	DefaultBufferSize = 100
	// DefaultMaxConcurrency is the default maximum number of concurrent requests
	DefaultMaxConcurrency = 2
)
