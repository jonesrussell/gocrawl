package transport

import "github.com/jonesrussell/gocrawl/internal/constants"

// Re-export transport constants from constants package for backward compatibility
const (
	DefaultMaxIdleConns          = constants.DefaultMaxIdleConns
	DefaultMaxIdleConnsPerHost   = constants.DefaultMaxIdleConnsPerHost
	DefaultIdleConnTimeout       = constants.DefaultIdleConnTimeout
	DefaultTLSHandshakeTimeout   = constants.DefaultTLSHandshakeTimeout
	DefaultResponseHeaderTimeout = constants.DefaultResponseHeaderTimeout
	DefaultExpectContinueTimeout = constants.DefaultExpectContinueTimeout
)
