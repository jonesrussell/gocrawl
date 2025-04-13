package common

import "time"

const (
	// DefaultCrawlerTimeout is the maximum time to wait for the crawler to complete.
	DefaultCrawlerTimeout = 30 * time.Minute
	// DefaultShutdownTimeout is the maximum time to wait for graceful shutdown.
	DefaultShutdownTimeout = 30 * time.Second
)
