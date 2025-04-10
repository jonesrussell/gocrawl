package crawler

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/types"
)

// Default configuration values
const (
	DefaultMaxDepth    = 3
	DefaultRateLimit   = 2 * time.Second
	DefaultParallelism = 2
	DefaultUserAgent   = "gocrawl/1.0"
	DefaultTimeout     = 30 * time.Second
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10MB
	DefaultRandomDelay = 500 * time.Millisecond
)

// Config represents the crawler configuration.
type Config struct {
	// BaseURL is the starting point for the crawler
	BaseURL string `yaml:"base_url"`
	// MaxDepth defines how many levels deep the crawler should traverse
	MaxDepth int `yaml:"max_depth"`
	// RateLimit defines the delay between requests
	RateLimit time.Duration `yaml:"rate_limit"`
	// RandomDelay adds randomization to the delay between requests
	RandomDelay time.Duration `yaml:"random_delay"`
	// Parallelism defines the number of concurrent crawlers
	Parallelism int `yaml:"parallelism"`
	// UserAgent defines the user agent string to use
	UserAgent string `yaml:"user_agent"`
	// Timeout defines the request timeout
	Timeout time.Duration `yaml:"timeout"`
	// MaxBodySize defines the maximum response body size
	MaxBodySize int64 `yaml:"max_body_size"`
	// AllowedDomains defines the domains that can be crawled
	AllowedDomains []string `yaml:"allowed_domains"`
	// DisallowedDomains defines the domains that cannot be crawled
	DisallowedDomains []string `yaml:"disallowed_domains"`
	// SourceFile is the path to the sources configuration file
	SourceFile string `yaml:"source_file"`
	// Sources contains the list of sources to crawl
	Sources []types.Source `yaml:"sources"`
	// IndexName is the name of the Elasticsearch index
	IndexName string `yaml:"index_name"`
}

// Validate validates the crawler configuration.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("crawler configuration is required")
	}

	if c.BaseURL == "" {
		return errors.New("base URL is required")
	}

	if c.MaxDepth < 0 {
		return errors.New("max depth must be non-negative")
	}

	if c.RateLimit < 0 {
		return errors.New("rate limit must be non-negative")
	}

	if c.RandomDelay < 0 {
		return errors.New("random delay must be non-negative")
	}

	if c.Parallelism < 1 {
		return errors.New("parallelism must be positive")
	}

	if c.Timeout < 0 {
		return errors.New("timeout must be non-negative")
	}

	if c.MaxBodySize < 0 {
		return errors.New("max body size must be non-negative")
	}

	return nil
}

// New creates a new crawler configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		MaxDepth:    DefaultMaxDepth,
		RateLimit:   DefaultRateLimit,
		RandomDelay: DefaultRandomDelay,
		Parallelism: DefaultParallelism,
		UserAgent:   DefaultUserAgent,
		Timeout:     DefaultTimeout,
		MaxBodySize: DefaultMaxBodySize,
		IndexName:   "gocrawl",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures a crawler configuration.
type Option func(*Config)

// WithBaseURL sets the base URL.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithMaxDepth sets the maximum depth.
func WithMaxDepth(depth int) Option {
	return func(c *Config) {
		c.MaxDepth = depth
	}
}

// WithRateLimit sets the rate limit.
func WithRateLimit(limit time.Duration) Option {
	return func(c *Config) {
		c.RateLimit = limit
	}
}

// WithRandomDelay sets the random delay.
func WithRandomDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.RandomDelay = delay
	}
}

// WithParallelism sets the parallelism.
func WithParallelism(parallelism int) Option {
	return func(c *Config) {
		c.Parallelism = parallelism
	}
}

// WithUserAgent sets the user agent.
func WithUserAgent(agent string) Option {
	return func(c *Config) {
		c.UserAgent = agent
	}
}

// WithTimeout sets the timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithMaxBodySize sets the maximum body size.
func WithMaxBodySize(size int64) Option {
	return func(c *Config) {
		c.MaxBodySize = size
	}
}

// WithAllowedDomains sets the allowed domains.
func WithAllowedDomains(domains []string) Option {
	return func(c *Config) {
		c.AllowedDomains = domains
	}
}

// WithDisallowedDomains sets the disallowed domains.
func WithDisallowedDomains(domains []string) Option {
	return func(c *Config) {
		c.DisallowedDomains = domains
	}
}

// WithSourceFile sets the source file.
func WithSourceFile(file string) Option {
	return func(c *Config) {
		c.SourceFile = file
	}
}

// ParseRateLimit parses a rate limit string into a time.Duration.
func ParseRateLimit(limit string) (time.Duration, error) {
	if limit == "" {
		return 0, errors.New("rate limit cannot be empty")
	}

	duration, err := time.ParseDuration(limit)
	if err != nil {
		return 0, fmt.Errorf("error parsing duration: %w", err)
	}

	if duration <= 0 {
		return 0, errors.New("rate limit must be positive")
	}

	return duration, nil
}
