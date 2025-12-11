// Package crawler provides configuration management for the web crawler component.
// It handles loading, validation, and access to crawler-specific settings
// such as crawl depth, concurrency, rate limiting, and TLS configuration.
package crawler

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Default configuration values
const (
	DefaultMaxDepth    = 5
	DefaultRateLimit   = 1 * time.Second
	DefaultParallelism = 5
	DefaultUserAgent   = "gocrawl/1.0"
	DefaultTimeout     = 30 * time.Second
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10MB
	DefaultRandomDelay = 500 * time.Millisecond
	DefaultMaxRetries  = 3
	DefaultRetryDelay  = 1 * time.Second
	// DefaultMaxRedirects is the default maximum number of redirects to follow
	DefaultMaxRedirects = 5
	// DefaultCleanupInterval is the default interval for cleanup operations
	DefaultCleanupInterval = 24 * time.Hour
)

// Config represents the crawler configuration.
type Config struct {
	// MaxDepth is the maximum depth to crawl
	MaxDepth int `yaml:"max_depth"`
	// MaxConcurrency is the maximum number of concurrent requests
	MaxConcurrency int `yaml:"max_concurrency"`
	// RequestTimeout is the timeout for each request
	RequestTimeout time.Duration `yaml:"request_timeout"`
	// UserAgent is the user agent to use for requests
	UserAgent string `yaml:"user_agent"`
	// RespectRobotsTxt indicates whether to respect robots.txt
	RespectRobotsTxt bool `yaml:"respect_robots_txt"`
	// AllowedDomains is the list of domains to crawl
	AllowedDomains []string `yaml:"allowed_domains"`
	// DisallowedDomains is the list of domains to exclude
	DisallowedDomains []string `yaml:"disallowed_domains"`
	// Delay is the delay between requests
	Delay time.Duration `yaml:"delay"`
	// RandomDelay is the random delay to add to the base delay
	RandomDelay time.Duration `yaml:"random_delay"`
	// SourcesAPIURL is the URL of the gosources API service
	SourcesAPIURL string `yaml:"sources_api_url"`
	// Debug enables debug logging
	Debug bool `yaml:"debug"`
	// TLS contains TLS configuration
	TLS TLSConfig `yaml:"tls"`
	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int `yaml:"max_retries"`
	// RetryDelay is the delay between retries
	RetryDelay time.Duration `yaml:"retry_delay"`
	// FollowRedirects indicates whether to follow redirects
	FollowRedirects bool `yaml:"follow_redirects"`
	// MaxRedirects is the maximum number of redirects to follow
	MaxRedirects int `yaml:"max_redirects"`
	// ValidateURLs indicates whether to validate URLs before visiting
	ValidateURLs bool `yaml:"validate_urls"`
	// CleanupInterval is the interval for cleaning up resources
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// Validate validates the crawler configuration.
func (c *Config) Validate() error {
	if c.MaxDepth < 0 {
		return errors.New("max_depth must be non-negative")
	}
	if c.MaxConcurrency < 1 {
		return errors.New("max_concurrency must be positive")
	}
	if c.RequestTimeout < 0 {
		return errors.New("request_timeout must be non-negative")
	}
	if c.Delay < 0 {
		return errors.New("delay must be non-negative")
	}
	if c.RandomDelay < 0 {
		return errors.New("random_delay must be non-negative")
	}
	return c.TLS.Validate()
}

// New creates a new crawler configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		MaxDepth:          DefaultMaxDepth,
		MaxConcurrency:    DefaultParallelism,
		RequestTimeout:    DefaultTimeout,
		Delay:             DefaultRateLimit,
		RandomDelay:       DefaultRandomDelay,
		UserAgent:         DefaultUserAgent,
		RespectRobotsTxt:  true,
		AllowedDomains:    []string{"*"},
		DisallowedDomains: []string{},
		SourcesAPIURL:     "http://localhost:8050/api/v1/sources",
		Debug:             false,
		TLS: TLSConfig{
			InsecureSkipVerify:       false, // Default to secure TLS verification
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               0, // Use highest supported version
			PreferServerCipherSuites: true,
		},
		MaxRetries:      DefaultMaxRetries,
		RetryDelay:      DefaultRetryDelay,
		FollowRedirects: true,
		MaxRedirects:    DefaultMaxRedirects,
		ValidateURLs:    true,
		CleanupInterval: DefaultCleanupInterval,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures a crawler configuration.
type Option func(*Config)

// WithMaxDepth sets the maximum depth.
func WithMaxDepth(depth int) Option {
	return func(c *Config) {
		c.MaxDepth = depth
	}
}

// WithMaxConcurrency sets the maximum concurrency.
func WithMaxConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.MaxConcurrency = concurrency
	}
}

// WithRequestTimeout sets the request timeout.
func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.RequestTimeout = timeout
	}
}

// WithUserAgent sets the user agent.
func WithUserAgent(agent string) Option {
	return func(c *Config) {
		c.UserAgent = agent
	}
}

// WithRespectRobotsTxt sets whether to respect robots.txt.
func WithRespectRobotsTxt(respect bool) Option {
	return func(c *Config) {
		c.RespectRobotsTxt = respect
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

// WithDelay sets the delay between requests.
func WithDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.Delay = delay
	}
}

// WithRandomDelay sets the random delay.
func WithRandomDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.RandomDelay = delay
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

// TLSConfig holds TLS-related configuration settings.
type TLSConfig struct {
	// InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
	// If InsecureSkipVerify is true, TLS accepts any certificate presented by the server
	// and any host name in that certificate.
	// In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used
	// only for testing or with trusted sources.
	// Default: false
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`

	// MinVersion is the minimum TLS version that is acceptable.
	// Default: TLS 1.2
	MinVersion uint16 `yaml:"min_version"`

	// MaxVersion is the maximum TLS version that is acceptable.
	// If zero, the maximum version supported by this package is used, which is currently TLS 1.3.
	// Default: 0 (use highest supported version)
	MaxVersion uint16 `yaml:"max_version"`

	// PreferServerCipherSuites controls whether the server's preference for cipher suites is honored.
	// If true, the server's preference is used. If false, the client's preference is used.
	// Default: true
	PreferServerCipherSuites bool `yaml:"prefer_server_cipher_suites"`
}

// NewTLSConfig creates a new TLS configuration with secure defaults.
func NewTLSConfig() *TLSConfig {
	return &TLSConfig{
		InsecureSkipVerify:       false,
		MinVersion:               tls.VersionTLS12,
		MaxVersion:               0, // Use highest supported version
		PreferServerCipherSuites: true,
	}
}

// Validate validates the TLS configuration.
func (c *TLSConfig) Validate() error {
	if c.InsecureSkipVerify {
		return errors.New("insecure_skip_verify is enabled. This is not recommended for " +
			"production use as it makes HTTPS connections vulnerable to man-in-the-middle attacks")
	}
	return nil
}

// LoadFromViper loads crawler configuration from Viper
// Viper will return values from (in order of precedence):
// 1. Environment variables
// 2. Config file
// 3. Defaults set via viper.SetDefault()
func LoadFromViper(v *viper.Viper) *Config {
	cfg := New()

	// Load values from Viper (will use defaults if not set)
	// Note: parallelism is an alias for max_concurrency
	if v.IsSet("crawler.parallelism") {
		cfg.MaxConcurrency = v.GetInt("crawler.parallelism")
	} else if v.IsSet("crawler.max_concurrency") {
		cfg.MaxConcurrency = v.GetInt("crawler.max_concurrency")
	}

	cfg.MaxDepth = v.GetInt("crawler.max_depth")
	cfg.RequestTimeout = v.GetDuration("crawler.request_timeout")
	cfg.UserAgent = v.GetString("crawler.user_agent")
	cfg.RespectRobotsTxt = v.GetBool("crawler.respect_robots_txt")
	cfg.AllowedDomains = v.GetStringSlice("crawler.allowed_domains")
	cfg.DisallowedDomains = v.GetStringSlice("crawler.disallowed_domains")
	cfg.Delay = v.GetDuration("crawler.delay")
	cfg.RandomDelay = v.GetDuration("crawler.random_delay")
	cfg.SourcesAPIURL = v.GetString("crawler.sources_api_url")
	cfg.Debug = v.GetBool("crawler.debug")
	// Only set MaxRetries if it's actually set and > 0 in Viper
	// This prevents overwriting the default with 0 when defaults haven't been loaded yet
	if maxRetries := v.GetInt("crawler.max_retries"); maxRetries > 0 {
		cfg.MaxRetries = maxRetries
	}
	// Only set RetryDelay if it's actually set and > 0 in Viper
	if retryDelay := v.GetDuration("crawler.retry_delay"); retryDelay > 0 {
		cfg.RetryDelay = retryDelay
	}
	cfg.FollowRedirects = v.GetBool("crawler.follow_redirects")
	cfg.MaxRedirects = v.GetInt("crawler.max_redirects")
	cfg.ValidateURLs = v.GetBool("crawler.validate_urls")
	// Only set CleanupInterval if it's actually set and non-zero in Viper
	// This prevents overwriting the default with 0 when defaults haven't been loaded yet
	if cleanupInterval := v.GetDuration("crawler.cleanup_interval"); cleanupInterval > 0 {
		cfg.CleanupInterval = cleanupInterval
	}

	// Load TLS configuration
	cfg.TLS.InsecureSkipVerify = v.GetBool("crawler.tls.insecure_skip_verify")
	if v.IsSet("crawler.tls.min_version") {
		minVer := v.GetInt("crawler.tls.min_version")
		if minVer >= 0 && minVer <= 65535 {
			cfg.TLS.MinVersion = uint16(minVer)
		}
	}
	if v.IsSet("crawler.tls.max_version") {
		maxVer := v.GetInt("crawler.tls.max_version")
		if maxVer >= 0 && maxVer <= 65535 {
			cfg.TLS.MaxVersion = uint16(maxVer)
		}
	}
	cfg.TLS.PreferServerCipherSuites = v.GetBool("crawler.tls.prefer_server_cipher_suites")

	return cfg
}
