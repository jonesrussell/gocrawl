package elasticsearch

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Default configuration values
const (
	DefaultAddresses     = "http://localhost:9200"
	DefaultIndexName     = "gocrawl"
	DefaultRetryEnabled  = true
	DefaultInitialWait   = 1 * time.Second
	DefaultMaxWait       = 5 * time.Second
	DefaultMaxRetries    = 3
	DefaultBulkSize      = 1000
	DefaultFlushInterval = 30 * time.Second
)

// Config holds Elasticsearch-specific configuration settings.
type Config struct {
	// Addresses is a comma-separated list of Elasticsearch nodes
	Addresses string `yaml:"addresses"`
	// IndexName is the name of the Elasticsearch index
	IndexName string `yaml:"index_name"`
	// APIKey is the API key for authentication (format: id:api_key)
	APIKey string `yaml:"api_key"`
	// RetryEnabled enables retry logic for failed requests
	RetryEnabled bool `yaml:"retry_enabled"`
	// InitialWait is the initial wait time between retries
	InitialWait time.Duration `yaml:"initial_wait"`
	// MaxWait is the maximum wait time between retries
	MaxWait time.Duration `yaml:"max_wait"`
	// MaxRetries is the maximum number of retries
	MaxRetries int `yaml:"max_retries"`
	// BulkSize is the number of documents to bulk index
	BulkSize int `yaml:"bulk_size"`
	// FlushInterval is the interval at which to flush the bulk indexer
	FlushInterval time.Duration `yaml:"flush_interval"`
	// TLSEnabled enables TLS for the connection
	TLSEnabled bool `yaml:"tls_enabled"`
	// TLSCertFile is the path to the TLS certificate file
	TLSCertFile string `yaml:"tls_cert_file"`
	// TLSKeyFile is the path to the TLS key file
	TLSKeyFile string `yaml:"tls_key_file"`
	// TLSCAFile is the path to the TLS CA file
	TLSCAFile string `yaml:"tls_ca_file"`
	// TLSInsecureSkipVerify skips TLS certificate verification
	TLSInsecureSkipVerify bool `yaml:"tls_insecure_skip_verify"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Addresses == "" {
		return errors.New("elasticsearch addresses cannot be empty")
	}

	if c.IndexName == "" {
		return errors.New("elasticsearch index name cannot be empty")
	}

	if c.APIKey == "" {
		return errors.New("elasticsearch API key cannot be empty")
	}

	if !strings.Contains(c.APIKey, ":") {
		return errors.New("elasticsearch API key must be in the format 'id:api_key'")
	}

	if c.RetryEnabled {
		if c.InitialWait <= 0 {
			return fmt.Errorf("initial wait must be greater than 0, got %v", c.InitialWait)
		}

		if c.MaxWait <= 0 {
			return fmt.Errorf("max wait must be greater than 0, got %v", c.MaxWait)
		}

		if c.MaxRetries <= 0 {
			return fmt.Errorf("max retries must be greater than 0, got %d", c.MaxRetries)
		}
	}

	if c.BulkSize <= 0 {
		return fmt.Errorf("bulk size must be greater than 0, got %d", c.BulkSize)
	}

	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush interval must be greater than 0, got %v", c.FlushInterval)
	}

	if c.TLSEnabled {
		if c.TLSCertFile == "" {
			return errors.New("TLS certificate file is required when TLS is enabled")
		}

		if c.TLSKeyFile == "" {
			return errors.New("TLS key file is required when TLS is enabled")
		}
	}

	return nil
}

// New creates a new Elasticsearch configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		Addresses:     DefaultAddresses,
		IndexName:     DefaultIndexName,
		RetryEnabled:  DefaultRetryEnabled,
		InitialWait:   DefaultInitialWait,
		MaxWait:       DefaultMaxWait,
		MaxRetries:    DefaultMaxRetries,
		BulkSize:      DefaultBulkSize,
		FlushInterval: DefaultFlushInterval,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures an Elasticsearch configuration.
type Option func(*Config)

// WithAddresses sets the Elasticsearch addresses.
func WithAddresses(addresses string) Option {
	return func(c *Config) {
		c.Addresses = addresses
	}
}

// WithIndexName sets the Elasticsearch index name.
func WithIndexName(name string) Option {
	return func(c *Config) {
		c.IndexName = name
	}
}

// WithAPIKey sets the Elasticsearch API key.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithRetryEnabled sets whether retry is enabled.
func WithRetryEnabled(enabled bool) Option {
	return func(c *Config) {
		c.RetryEnabled = enabled
	}
}

// WithInitialWait sets the initial wait time.
func WithInitialWait(wait time.Duration) Option {
	return func(c *Config) {
		c.InitialWait = wait
	}
}

// WithMaxWait sets the maximum wait time.
func WithMaxWait(wait time.Duration) Option {
	return func(c *Config) {
		c.MaxWait = wait
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithBulkSize sets the bulk size.
func WithBulkSize(size int) Option {
	return func(c *Config) {
		c.BulkSize = size
	}
}

// WithFlushInterval sets the flush interval.
func WithFlushInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.FlushInterval = interval
	}
}

// WithTLSEnabled sets whether TLS is enabled.
func WithTLSEnabled(enabled bool) Option {
	return func(c *Config) {
		c.TLSEnabled = enabled
	}
}

// WithTLSCertFile sets the TLS certificate file.
func WithTLSCertFile(file string) Option {
	return func(c *Config) {
		c.TLSCertFile = file
	}
}

// WithTLSKeyFile sets the TLS key file.
func WithTLSKeyFile(file string) Option {
	return func(c *Config) {
		c.TLSKeyFile = file
	}
}

// WithTLSCAFile sets the TLS CA file.
func WithTLSCAFile(file string) Option {
	return func(c *Config) {
		c.TLSCAFile = file
	}
}

// WithTLSInsecureSkipVerify sets whether to skip TLS certificate verification.
func WithTLSInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		c.TLSInsecureSkipVerify = skip
	}
}
