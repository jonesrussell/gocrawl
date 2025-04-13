package sources

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Default configuration values
const (
	DefaultMaxDepth     = 2
	DefaultParallelism  = 2
	DefaultRateLimit    = 2 * time.Second
	DefaultTimeout      = 30 * time.Second
	DefaultUserAgent    = "gocrawl/1.0"
	DefaultAllowDomains = "*"
)

// Config holds source-specific configuration settings.
type Config struct {
	// BaseURL is the starting URL for the crawler
	BaseURL string `yaml:"base_url"`
	// MaxDepth is the maximum depth to crawl
	MaxDepth int `yaml:"max_depth"`
	// Parallelism is the number of concurrent requests
	Parallelism int `yaml:"parallelism"`
	// RateLimit is the delay between requests
	RateLimit time.Duration `yaml:"rate_limit"`
	// Timeout is the request timeout
	Timeout time.Duration `yaml:"timeout"`
	// UserAgent is the user agent string to use
	UserAgent string `yaml:"user_agent"`
	// AllowDomains is a comma-separated list of allowed domains
	AllowDomains string `yaml:"allow_domains"`
	// DisallowedURLFilters is a list of URL patterns to exclude
	DisallowedURLFilters []string `yaml:"disallowed_url_filters"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("base URL cannot be empty")
	}

	// Validate base URL format
	parsedURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	// Additional validation for URL components
	if parsedURL.Scheme == "" {
		return errors.New("base URL must include a scheme (http:// or https://)")
	}
	if parsedURL.Host == "" {
		return errors.New("base URL must include a host")
	}

	if c.MaxDepth < 0 {
		return errors.New("max depth must be greater than or equal to 0")
	}

	if c.Parallelism <= 0 {
		return errors.New("parallelism must be greater than 0")
	}

	if c.RateLimit < 0 {
		return errors.New("rate limit must be greater than or equal to 0")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.UserAgent == "" {
		return errors.New("user agent cannot be empty")
	}

	if c.AllowDomains == "" {
		return errors.New("allow domains cannot be empty")
	}

	return nil
}

// New creates a new sources configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		MaxDepth:     DefaultMaxDepth,
		Parallelism:  DefaultParallelism,
		RateLimit:    DefaultRateLimit,
		Timeout:      DefaultTimeout,
		UserAgent:    DefaultUserAgent,
		AllowDomains: DefaultAllowDomains,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures a sources configuration.
type Option func(*Config)

// WithBaseURL sets the base URL for the crawler.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithMaxDepth sets the maximum crawl depth.
func WithMaxDepth(depth int) Option {
	return func(c *Config) {
		c.MaxDepth = depth
	}
}

// WithParallelism sets the number of concurrent requests.
func WithParallelism(parallelism int) Option {
	return func(c *Config) {
		c.Parallelism = parallelism
	}
}

// WithRateLimit sets the delay between requests.
func WithRateLimit(limit time.Duration) Option {
	return func(c *Config) {
		c.RateLimit = limit
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithUserAgent sets the user agent string.
func WithUserAgent(agent string) Option {
	return func(c *Config) {
		c.UserAgent = agent
	}
}

// WithAllowDomains sets the allowed domains.
func WithAllowDomains(domains string) Option {
	return func(c *Config) {
		c.AllowDomains = domains
	}
}

// WithDisallowedURLFilters sets the URL patterns to exclude.
func WithDisallowedURLFilters(filters []string) Option {
	return func(c *Config) {
		c.DisallowedURLFilters = filters
	}
}
