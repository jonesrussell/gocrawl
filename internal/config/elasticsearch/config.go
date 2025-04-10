// Package elasticsearch provides Elasticsearch configuration management.
package elasticsearch

import (
	"errors"
	"fmt"
	"time"
)

// ValidationLevel represents the level of configuration validation required
type ValidationLevel int

const (
	// BasicValidation only validates essential connection settings
	BasicValidation ValidationLevel = iota
	// FullValidation performs complete configuration validation
	FullValidation
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
	MinPasswordLength    = 8
	DefaultDiscoverNodes = false // Default to false to prevent node discovery

	// Retry configuration constants
	HAInitialWait = 2 * time.Second
	HAMaxWait     = 10 * time.Second
	HAMaxRetries  = 5
)

// Error codes for configuration validation
const (
	ErrCodeEmptyAddresses  = "EMPTY_ADDRESSES"
	ErrCodeEmptyIndexName  = "EMPTY_INDEX_NAME"
	ErrCodeMissingAPIKey   = "MISSING_API_KEY"
	ErrCodeInvalidFormat   = "INVALID_FORMAT"
	ErrCodeWeakPassword    = "WEAK_PASSWORD"
	ErrCodeInvalidRetry    = "INVALID_RETRY"
	ErrCodeInvalidBulkSize = "INVALID_BULK_SIZE"
	ErrCodeInvalidFlush    = "INVALID_FLUSH"
	ErrCodeInvalidTLS      = "INVALID_TLS"
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	Code    string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Config represents Elasticsearch configuration settings.
type Config struct {
	// Addresses is a list of Elasticsearch node addresses
	Addresses []string `yaml:"addresses"`
	// APIKey is the base64 encoded API key for authentication
	APIKey string `yaml:"api_key"`
	// Username is the username for authentication
	Username string `yaml:"username"`
	// Password is the password for authentication (minimum 8 characters)
	Password string `yaml:"password"`
	// IndexName is the name of the index
	IndexName string `yaml:"index_name"`
	// Cloud contains cloud-specific configuration
	Cloud struct {
		ID     string `yaml:"id"`
		APIKey string `yaml:"api_key"`
	} `yaml:"cloud"`
	// TLS contains TLS configuration
	TLS *TLSConfig `yaml:"tls"`
	// Retry contains retry configuration
	Retry struct {
		Enabled     bool          `yaml:"enabled"`
		InitialWait time.Duration `yaml:"initial_wait"`
		MaxWait     time.Duration `yaml:"max_wait"`
		MaxRetries  int           `yaml:"max_retries"`
	} `yaml:"retry"`
	// BulkSize is the number of documents to bulk index
	BulkSize int `yaml:"bulk_size"`
	// FlushInterval is the interval at which to flush the bulk indexer
	FlushInterval time.Duration `yaml:"flush_interval"`
	// DiscoverNodes enables/disables node discovery
	DiscoverNodes bool `yaml:"discover_nodes"`
}

// TLSConfig represents TLS configuration settings.
type TLSConfig struct {
	CertFile           string `yaml:"cert_file" env:"ELASTICSEARCH_TLS_CERT_FILE"`
	KeyFile            string `yaml:"key_file" env:"ELASTICSEARCH_TLS_KEY_FILE"`
	CAFile             string `yaml:"ca_file" env:"ELASTICSEARCH_TLS_CA_FILE"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" env:"ELASTICSEARCH_TLS_INSECURE_SKIP_VERIFY"`
}

// Validate performs complete configuration validation
func (c *Config) Validate() error {
	if err := c.ValidateConnection(); err != nil {
		return err
	}

	if c.IndexName == "" {
		return errors.New("index name is required")
	}

	return nil
}

// ValidateConnection validates only the connection settings
func (c *Config) ValidateConnection() error {
	if len(c.Addresses) == 0 {
		return errors.New("at least one Elasticsearch address is required")
	}
	return nil
}

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Addresses: []string{DefaultAddresses},
		IndexName: DefaultIndexName,
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     DefaultRetryEnabled,
			InitialWait: DefaultInitialWait,
			MaxWait:     DefaultMaxWait,
			MaxRetries:  DefaultMaxRetries,
		},
		BulkSize:      DefaultBulkSize,
		FlushInterval: DefaultFlushInterval,
		TLS: &TLSConfig{
			InsecureSkipVerify: true, // Default to true for development
		},
	}
}

// ExampleConfig demonstrates common configuration patterns.
func ExampleConfig() {
	// Basic configuration
	basicCfg := NewConfig()
	WithAddresses([]string{"http://localhost:9200"})(basicCfg)
	WithAPIKey("my_id:my_key")(basicCfg)
	WithIndexName("my_index")(basicCfg)

	// Cloud configuration
	cloudCfg := NewConfig()
	WithCloudID("my-cloud-id")(cloudCfg)
	WithCloudAPIKey("my-cloud-key")(cloudCfg)

	// TLS configuration
	tlsCfg := NewConfig()
	WithTLSCertFile("cert.pem")(tlsCfg)
	WithTLSKeyFile("key.pem")(tlsCfg)
	WithTLSCAFile("ca.pem")(tlsCfg)

	// High availability configuration
	haCfg := NewConfig()
	WithAddresses([]string{
		"http://node1:9200",
		"http://node2:9200",
		"http://node3:9200",
	})(haCfg)
	WithAPIKey("ha_id:ha_key")(haCfg)
	WithIndexName("ha_index")(haCfg)
	WithRetryEnabled(true)(haCfg)
	WithInitialWait(HAInitialWait)(haCfg)
	WithMaxWait(HAMaxWait)(haCfg)
	WithMaxRetries(HAMaxRetries)(haCfg)

	// Validate configurations
	_ = basicCfg.Validate()
	_ = cloudCfg.Validate()
	_ = tlsCfg.Validate()
	_ = haCfg.Validate()
}

// Option is a function that configures an Elasticsearch configuration.
type Option func(*Config)

// WithAddresses sets the Elasticsearch addresses.
func WithAddresses(addresses []string) Option {
	return func(c *Config) {
		c.Addresses = addresses
	}
}

// WithAPIKey sets the Elasticsearch API key.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithUsername sets the Elasticsearch username.
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword sets the Elasticsearch password.
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithIndexName sets the Elasticsearch index name.
func WithIndexName(name string) Option {
	return func(c *Config) {
		c.IndexName = name
	}
}

// WithCloudID sets the Elasticsearch cloud ID.
func WithCloudID(id string) Option {
	return func(c *Config) {
		c.Cloud.ID = id
	}
}

// WithCloudAPIKey sets the Elasticsearch cloud API key.
func WithCloudAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.Cloud.APIKey = apiKey
	}
}

// WithRetryEnabled sets whether retry is enabled.
func WithRetryEnabled(enabled bool) Option {
	return func(c *Config) {
		c.Retry.Enabled = enabled
	}
}

// WithInitialWait sets the initial wait time.
func WithInitialWait(wait time.Duration) Option {
	return func(c *Config) {
		c.Retry.InitialWait = wait
	}
}

// WithMaxWait sets the maximum wait time.
func WithMaxWait(wait time.Duration) Option {
	return func(c *Config) {
		c.Retry.MaxWait = wait
	}
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.Retry.MaxRetries = retries
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

// WithTLSCertFile sets the TLS certificate file.
func WithTLSCertFile(file string) Option {
	return func(c *Config) {
		if c.TLS == nil {
			c.TLS = &TLSConfig{}
		}
		c.TLS.CertFile = file
	}
}

// WithTLSKeyFile sets the TLS key file.
func WithTLSKeyFile(file string) Option {
	return func(c *Config) {
		if c.TLS == nil {
			c.TLS = &TLSConfig{}
		}
		c.TLS.KeyFile = file
	}
}

// WithTLSCAFile sets the TLS CA file.
func WithTLSCAFile(file string) Option {
	return func(c *Config) {
		if c.TLS == nil {
			c.TLS = &TLSConfig{}
		}
		c.TLS.CAFile = file
	}
}

// WithTLSInsecureSkipVerify sets whether to skip TLS certificate verification.
func WithTLSInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		if c.TLS == nil {
			c.TLS = &TLSConfig{}
		}
		c.TLS.InsecureSkipVerify = skip
	}
}
