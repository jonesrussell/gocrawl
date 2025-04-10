package elasticsearch

import (
	"errors"
	"fmt"
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

// Config represents Elasticsearch configuration settings.
type Config struct {
	// Addresses is a list of Elasticsearch node addresses
	Addresses []string `yaml:"addresses"`
	// APIKey is the API key for authentication
	APIKey string `yaml:"api_key"`
	// Username is the username for authentication
	Username string `yaml:"username"`
	// Password is the password for authentication
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
}

// TLSConfig represents TLS configuration settings.
type TLSConfig struct {
	// Enabled indicates whether TLS is enabled
	Enabled bool `yaml:"enabled"`
	// CertFile is the path to the certificate file
	CertFile string `yaml:"cert_file"`
	// KeyFile is the path to the key file
	KeyFile string `yaml:"key_file"`
	// CAFile is the path to the CA certificate file
	CAFile string `yaml:"ca_file"`
	// InsecureSkipVerify indicates whether to skip certificate verification
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("elasticsearch addresses cannot be empty")
	}

	if c.IndexName == "" {
		return errors.New("elasticsearch index name cannot be empty")
	}

	if c.APIKey == "" {
		return errors.New("elasticsearch API key cannot be empty")
	}

	if c.Retry.Enabled {
		if c.Retry.InitialWait <= 0 {
			return fmt.Errorf("initial wait must be greater than 0, got %v", c.Retry.InitialWait)
		}

		if c.Retry.MaxWait <= 0 {
			return fmt.Errorf("max wait must be greater than 0, got %v", c.Retry.MaxWait)
		}

		if c.Retry.MaxRetries <= 0 {
			return fmt.Errorf("max retries must be greater than 0, got %d", c.Retry.MaxRetries)
		}
	}

	if c.BulkSize <= 0 {
		return fmt.Errorf("bulk size must be greater than 0, got %d", c.BulkSize)
	}

	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush interval must be greater than 0, got %v", c.FlushInterval)
	}

	if c.TLS.Enabled {
		if !c.TLS.InsecureSkipVerify {
			if c.TLS.CertFile == "" {
				return errors.New("TLS certificate file is required when TLS is enabled and insecure_skip_verify is false")
			}

			if c.TLS.KeyFile == "" {
				return errors.New("TLS key file is required when TLS is enabled and insecure_skip_verify is false")
			}
		}
		// When insecure_skip_verify is true, we don't need certificate files
	}

	return nil
}

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	defaultAddresses := []string{DefaultAddresses}
	tls := &TLSConfig{
		Enabled:            false,
		InsecureSkipVerify: true,
	}

	return &Config{
		Addresses: defaultAddresses,
		Cloud: struct {
			ID     string `yaml:"id"`
			APIKey string `yaml:"api_key"`
		}{},
		TLS: tls,
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
	}
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

// WithTLSEnabled sets whether TLS is enabled.
func WithTLSEnabled(enabled bool) Option {
	return func(c *Config) {
		c.TLS.Enabled = enabled
	}
}

// WithTLSCertFile sets the TLS certificate file.
func WithTLSCertFile(file string) Option {
	return func(c *Config) {
		c.TLS.CertFile = file
	}
}

// WithTLSKeyFile sets the TLS key file.
func WithTLSKeyFile(file string) Option {
	return func(c *Config) {
		c.TLS.KeyFile = file
	}
}

// WithTLSCAFile sets the TLS CA file.
func WithTLSCAFile(file string) Option {
	return func(c *Config) {
		c.TLS.CAFile = file
	}
}

// WithTLSInsecureSkipVerify sets whether to skip TLS certificate verification.
func WithTLSInsecureSkipVerify(skip bool) Option {
	return func(c *Config) {
		c.TLS.InsecureSkipVerify = skip
	}
}
