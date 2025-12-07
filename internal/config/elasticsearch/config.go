// Package elasticsearch provides Elasticsearch configuration management.
package elasticsearch

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
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
	DefaultAddresses     = "http://127.0.0.1:9200"
	DefaultIndexName     = "gocrawl"
	DefaultRetryEnabled  = true
	DefaultInitialWait   = 1 * time.Second
	DefaultMaxWait       = 5 * time.Second
	DefaultMaxRetries    = 3
	DefaultBulkSize      = 1000
	DefaultFlushInterval = 30 * time.Second
	MinPasswordLength    = 8
	DefaultDiscoverNodes = false // Default to false to prevent node discovery
	DefaultUsername      = "elastic"
	DefaultPassword      = "changeme"
	DefaultMaxSize       = 1024 * 1024 * 1024 // 1 GB
	DefaultMaxItems      = 10000

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
	Addresses []string `yaml:"addresses" env:"ELASTICSEARCH_HOSTS"`
	// APIKey is the base64 encoded API key for authentication
	APIKey string `yaml:"api_key" env:"ELASTICSEARCH_API_KEY"`
	// Username is the username for authentication
	Username string `yaml:"username" env:"ELASTICSEARCH_USERNAME"`
	// Password is the password for authentication (minimum 8 characters)
	Password string `yaml:"password" env:"ELASTICSEARCH_PASSWORD"`
	// IndexName is the name of the index
	IndexName string `yaml:"index_name" env:"ELASTICSEARCH_INDEX_PREFIX"`
	// Cloud contains cloud-specific configuration
	Cloud struct {
		ID     string `yaml:"id"`
		APIKey string `yaml:"api_key"`
	} `yaml:"cloud"`
	// TLS contains TLS configuration
	TLS *TLSConfig `yaml:"tls"`
	// Retry contains retry configuration
	Retry struct {
		Enabled     bool          `yaml:"enabled" env:"ELASTICSEARCH_RETRY_ENABLED"`
		InitialWait time.Duration `yaml:"initial_wait" env:"ELASTICSEARCH_RETRY_INITIAL_WAIT"`
		MaxWait     time.Duration `yaml:"max_wait" env:"ELASTICSEARCH_RETRY_MAX_WAIT"`
		MaxRetries  int           `yaml:"max_retries" env:"ELASTICSEARCH_MAX_RETRIES"`
	} `yaml:"retry"`
	// BulkSize is the number of documents to bulk index
	BulkSize int `yaml:"bulk_size"`
	// FlushInterval is the interval at which to flush the bulk indexer
	FlushInterval time.Duration `yaml:"flush_interval"`
	// DiscoverNodes enables/disables node discovery
	DiscoverNodes bool `yaml:"discover_nodes" env:"ELASTICSEARCH_DISCOVER_NODES"`
	// MaxSize is the maximum size of the storage in bytes
	MaxSize int64 `yaml:"max_size"`
	// MaxItems is the maximum number of items to store
	MaxItems int `yaml:"max_items"`
	// Compression is whether to compress stored data
	Compression bool `yaml:"compression"`
}

// TLSConfig represents TLS configuration settings.
type TLSConfig struct {
	CertFile           string `yaml:"cert_file" env:"ELASTICSEARCH_TLS_CERT_FILE"`
	KeyFile            string `yaml:"key_file" env:"ELASTICSEARCH_TLS_KEY_FILE"`
	CAFile             string `yaml:"ca_file" env:"ELASTICSEARCH_TLS_CA_FILE"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" env:"ELASTICSEARCH_TLS_INSECURE_SKIP_VERIFY"`
	Enabled            bool   `yaml:"enabled" env:"ELASTICSEARCH_TLS_ENABLED"`
}

// validateTLS validates the TLS configuration
func (c *Config) validateTLS() error {
	if c.TLS != nil {
		if (c.TLS.CertFile != "" && c.TLS.KeyFile == "") || (c.TLS.CertFile == "" && c.TLS.KeyFile != "") {
			return &ConfigError{
				Code:    ErrCodeInvalidTLS,
				Message: "both cert file and key file must be provided for TLS",
			}
		}
	}
	return nil
}

// validateRequiredFields validates required configuration fields
func (c *Config) validateRequiredFields() error {
	if len(c.Addresses) == 0 {
		return &ConfigError{
			Code:    ErrCodeEmptyAddresses,
			Message: "at least one address is required",
		}
	}

	if c.IndexName == "" {
		return &ConfigError{
			Code:    ErrCodeEmptyIndexName,
			Message: "index name is required",
		}
	}

	// Allow either API key or username/password authentication
	if c.APIKey == "" && (c.Username == "" || c.Password == "") {
		return &ConfigError{
			Code:    ErrCodeMissingAPIKey,
			Message: "either API key or username/password is required",
		}
	}

	return nil
}

// validatePassword validates the password configuration
func (c *Config) validatePassword() error {
	if c.Password != "" && len(c.Password) < MinPasswordLength {
		return &ConfigError{
			Code:    ErrCodeWeakPassword,
			Message: fmt.Sprintf("password must be at least %d characters", MinPasswordLength),
		}
	}
	return nil
}

// validateRetry validates the retry configuration
func (c *Config) validateRetry() error {
	if c.Retry.Enabled {
		if c.Retry.InitialWait < 0 || c.Retry.MaxWait < 0 || c.Retry.MaxRetries < 0 {
			return &ConfigError{
				Code:    ErrCodeInvalidRetry,
				Message: "retry configuration must be non-negative",
			}
		}
	}
	return nil
}

// validateBulkConfig validates bulk indexing configuration
func (c *Config) validateBulkConfig() error {
	if c.FlushInterval <= 0 {
		return &ConfigError{
			Code:    ErrCodeInvalidFlush,
			Message: "flush interval must be positive",
		}
	}

	if c.BulkSize <= 0 {
		return &ConfigError{
			Code:    ErrCodeInvalidBulkSize,
			Message: "bulk size must be positive",
		}
	}

	return nil
}

// validateAPIKeyFormat validates the API key format
func (c *Config) validateAPIKeyFormat() error {
	if c.APIKey != "" {
		parts := strings.Split(c.APIKey, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return &ConfigError{
				Code:    ErrCodeInvalidFormat,
				Message: "API key must be in the format 'id:key'",
			}
		}
	}
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c == nil {
		return &ConfigError{
			Code:    ErrCodeEmptyAddresses,
			Message: "configuration is required",
		}
	}

	// Validate each component
	if err := c.validateTLS(); err != nil {
		return err
	}

	if err := c.validateRequiredFields(); err != nil {
		return err
	}

	if err := c.validatePassword(); err != nil {
		return err
	}

	if err := c.validateRetry(); err != nil {
		return err
	}

	if err := c.validateBulkConfig(); err != nil {
		return err
	}

	if err := c.validateAPIKeyFormat(); err != nil {
		return err
	}

	return nil
}

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		Addresses: []string{DefaultAddresses},
		Retry: struct {
			Enabled     bool          `yaml:"enabled" env:"ELASTICSEARCH_RETRY_ENABLED"`
			InitialWait time.Duration `yaml:"initial_wait" env:"ELASTICSEARCH_RETRY_INITIAL_WAIT"`
			MaxWait     time.Duration `yaml:"max_wait" env:"ELASTICSEARCH_RETRY_MAX_WAIT"`
			MaxRetries  int           `yaml:"max_retries" env:"ELASTICSEARCH_MAX_RETRIES"`
		}{
			Enabled:     DefaultRetryEnabled,
			InitialWait: DefaultInitialWait,
			MaxWait:     DefaultMaxWait,
			MaxRetries:  DefaultMaxRetries,
		},
		BulkSize:      DefaultBulkSize,
		FlushInterval: DefaultFlushInterval,
		TLS: &TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
		},
		MaxSize:     DefaultMaxSize,
		MaxItems:    DefaultMaxItems,
		Compression: true,
	}
}

// Option is a function that configures an Elasticsearch configuration.
type Option func(*Config)

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

// parseAddressesFromString parses comma-separated addresses from a string
func parseAddressesFromString(addrStr string) []string {
	addresses := strings.Split(addrStr, ",")
	// Trim whitespace from each address
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
	}
	// Remove any empty strings after trimming
	var filtered []string
	for _, addr := range addresses {
		if addr != "" {
			filtered = append(filtered, addr)
		}
	}
	return filtered
}

// parseAddressesFromConfig parses addresses from Viper configuration
func parseAddressesFromConfig(v *viper.Viper) []string {
	// Try GetStringSlice first (for YAML configs with array syntax)
	addresses := v.GetStringSlice("elasticsearch.addresses")
	// Filter out empty strings
	var filtered []string
	for _, addr := range addresses {
		if addr != "" {
			filtered = append(filtered, addr)
		}
	}
	addresses = filtered

	// If GetStringSlice returned empty, fall back to GetString and parse as comma-separated
	if len(addresses) == 0 {
		addrStr := v.GetString("elasticsearch.addresses")
		if addrStr != "" {
			addresses = parseAddressesFromString(addrStr)
		} else {
			// Fallback to single string value
			if addr := v.GetString("elasticsearch.address"); addr != "" {
				addresses = []string{addr}
			}
		}
	}
	return addresses
}

// LoadFromViper loads Elasticsearch configuration from Viper
func LoadFromViper(v *viper.Viper) *Config {
	var addresses []string

	// First, check environment variables directly to ensure they take precedence over defaults
	// Environment variables are always comma-separated strings
	var addrStr string
	if envAddr := os.Getenv("ELASTICSEARCH_HOSTS"); envAddr != "" {
		addrStr = envAddr
	} else if envAddr2 := os.Getenv("ELASTICSEARCH_ADDRESSES"); envAddr2 != "" {
		addrStr = envAddr2
	}

	// If we have an environment variable, parse it as a comma-separated string
	if addrStr != "" {
		addresses = parseAddressesFromString(addrStr)
	} else {
		addresses = parseAddressesFromConfig(v)
	}

	// If addresses is still empty, use default
	if len(addresses) == 0 {
		addresses = []string{DefaultAddresses}
	}

	// Get index name - support both index_name and index_prefix for backward compatibility
	indexName := v.GetString("elasticsearch.index_name")
	if indexName == "" {
		indexName = v.GetString("elasticsearch.index_prefix")
	}
	if indexName == "" {
		indexName = DefaultIndexName
	}

	cfg := &Config{
		Addresses: addresses,
		IndexName: indexName,
		Username:  v.GetString("elasticsearch.username"),
		Password:  v.GetString("elasticsearch.password"),
		APIKey:    v.GetString("elasticsearch.api_key"),
		TLS: &TLSConfig{
			Enabled:            v.GetBool("elasticsearch.tls.enabled"),
			InsecureSkipVerify: v.GetBool("elasticsearch.tls.insecure_skip_verify"),
			CAFile:             v.GetString("elasticsearch.tls.ca_file"),
			CertFile:           v.GetString("elasticsearch.tls.cert_file"),
			KeyFile:            v.GetString("elasticsearch.tls.key_file"),
		},
	}
	return cfg
}
