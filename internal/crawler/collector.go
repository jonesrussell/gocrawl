package crawler

import (
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/internal/constants"
)

// CollectorConfig holds configuration for the collector.
type CollectorConfig struct {
	RateLimit      time.Duration
	MaxDepth       int
	MaxConcurrency int
}

// NewCollectorConfig creates a new collector configuration.
func NewCollectorConfig() *CollectorConfig {
	return &CollectorConfig{
		RateLimit:      constants.DefaultRateLimit,
		MaxDepth:       constants.DefaultMaxDepth,
		MaxConcurrency: constants.DefaultMaxConcurrency,
	}
}

// Validate validates the collector configuration.
func (c *CollectorConfig) Validate() error {
	if c.RateLimit < 0 {
		return errors.New("rate limit must be non-negative")
	}
	if c.MaxDepth < 0 {
		return errors.New("max depth must be non-negative")
	}
	if c.MaxConcurrency < 1 {
		return errors.New("max concurrency must be positive")
	}
	return nil
}
