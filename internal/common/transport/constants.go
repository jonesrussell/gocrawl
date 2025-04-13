package transport

import "time"

const (
	// DefaultMaxIdleConns is the default maximum number of idle connections
	DefaultMaxIdleConns = 100
	// DefaultMaxIdleConnsPerHost is the default maximum number of idle connections per host
	DefaultMaxIdleConnsPerHost = 10
	// DefaultIdleConnTimeout is the default timeout for idle connections
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultTLSHandshakeTimeout is the default timeout for TLS handshake
	DefaultTLSHandshakeTimeout = 10 * time.Second
	// DefaultResponseHeaderTimeout is the default timeout for response headers
	DefaultResponseHeaderTimeout = 30 * time.Second
	// DefaultExpectContinueTimeout is the default timeout for expect continue
	DefaultExpectContinueTimeout = 1 * time.Second
)
