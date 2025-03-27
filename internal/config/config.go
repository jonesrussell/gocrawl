// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Constants for default configuration values
const (
	defaultRateLimit = 2 * time.Second
)

// AppConfig holds application-level configuration settings.
// It contains basic information about the application instance.
type AppConfig struct {
	// Environment specifies the runtime environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Name is the application name
	Name string `yaml:"name"`
	// Version is the application version
	Version string `yaml:"version"`
	// Debug enables debug mode for additional logging
	Debug bool `yaml:"debug"`
}

// CrawlerConfig holds crawler-specific configuration settings.
// It defines how the crawler should behave when collecting content.
type CrawlerConfig struct {
	// BaseURL is the starting point for the crawler
	BaseURL string `yaml:"base_url"`
	// MaxDepth defines how many levels deep the crawler should traverse
	MaxDepth int `yaml:"max_depth"`
	// RateLimit defines the delay between requests
	RateLimit time.Duration `yaml:"rate_limit"`
	// RandomDelay adds randomization to the delay between requests
	RandomDelay time.Duration `yaml:"random_delay"`
	// IndexName is the Elasticsearch index for storing crawled content
	IndexName string `yaml:"index_name"`
	// ContentIndexName is the Elasticsearch index for storing parsed content
	ContentIndexName string `yaml:"content_index_name"`
	// SourceFile is the path to the sources configuration file
	SourceFile string `yaml:"source_file"`
	// Parallelism defines how many concurrent crawlers to run
	Parallelism int `yaml:"parallelism"`
}

// SetMaxDepth sets the MaxDepth in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - depth: The maximum depth for crawling
func (c *CrawlerConfig) SetMaxDepth(depth int) {
	c.MaxDepth = depth
	viper.Set("crawler.max_depth", depth)
}

// SetRateLimit sets the RateLimit in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - rate: The time duration between requests
func (c *CrawlerConfig) SetRateLimit(rate time.Duration) {
	c.RateLimit = rate
	viper.Set("crawler.rate_limit", rate.String())
}

// SetBaseURL sets the BaseURL in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - url: The base URL for crawling
func (c *CrawlerConfig) SetBaseURL(url string) {
	c.BaseURL = url
	viper.Set("crawler.base_url", url)
}

// SetIndexName sets the IndexName in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - index: The name of the Elasticsearch index
func (c *CrawlerConfig) SetIndexName(index string) {
	c.IndexName = index
	viper.Set("elasticsearch.index_name", index)
}

// ElasticsearchConfig contains Elasticsearch connection and configuration settings.
type ElasticsearchConfig struct {
	// Addresses is a list of Elasticsearch node addresses
	Addresses []string `yaml:"addresses"`
	// Username for Elasticsearch authentication
	Username string `yaml:"username"`
	// Password for Elasticsearch authentication
	Password string `yaml:"password"`
	// APIKey for Elasticsearch authentication
	APIKey string `yaml:"api_key"`
	// IndexName is the default index for storing data
	IndexName string `yaml:"index_name"`
	// Cloud configuration for Elastic Cloud
	Cloud struct {
		// ID is the Elastic Cloud deployment ID
		ID string `yaml:"id"`
		// APIKey is the Elastic Cloud API key
		APIKey string `yaml:"api_key"`
	} `yaml:"cloud"`
	// TLS configuration
	TLS struct {
		// Enabled indicates whether to use TLS
		Enabled bool `yaml:"enabled"`
		// SkipVerify indicates whether to skip TLS certificate verification
		SkipVerify bool `yaml:"skip_verify"`
		// Certificate is the path to the client certificate
		Certificate string `yaml:"certificate"`
		// Key is the path to the client key
		Key string `yaml:"key"`
		// CA is the path to the CA certificate
		CA string `yaml:"ca"`
	} `yaml:"tls"`
	// Retry configuration
	Retry struct {
		// Enabled indicates whether to enable retries
		Enabled bool `yaml:"enabled"`
		// InitialWait is the initial wait time between retries
		InitialWait time.Duration `yaml:"initial_wait"`
		// MaxWait is the maximum wait time between retries
		MaxWait time.Duration `yaml:"max_wait"`
		// MaxRetries is the maximum number of retries
		MaxRetries int `yaml:"max_retries"`
	} `yaml:"retry"`
}

// LogConfig holds logging-related configuration settings.
// It defines how the application should handle logging.
type LogConfig struct {
	// Level defines the minimum level of logs to output
	Level string `yaml:"level"`
	// Debug enables additional debug logging
	Debug bool `yaml:"debug"`
}

// Source represents a news source configuration.
// It defines how to crawl and process content from a specific source.
type Source struct {
	// Name is the unique identifier for the source
	Name string `yaml:"name"`
	// URL is the starting point for crawling this source
	URL string `yaml:"url"`
	// RateLimit defines the delay between requests for this source
	RateLimit time.Duration `yaml:"rate_limit"`
	// MaxDepth defines how deep to crawl this source
	MaxDepth int `yaml:"max_depth"`
	// ArticleIndex is the Elasticsearch index for storing articles
	ArticleIndex string `yaml:"article_index"`
	// Index is the Elasticsearch index for storing non-article content
	Index string `yaml:"index"`
	// Time defines which time-related fields to extract
	Time []string `yaml:"time,omitempty"`
	// Selectors defines how to extract content from pages
	Selectors SourceSelectors `yaml:"selectors"`
	// Metadata contains additional metadata for the source
	Metadata map[string]string `yaml:"metadata,omitempty"`
}

// SourceSelectors defines the selectors for a source.
// It contains the rules for extracting content from web pages.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
}

// ServerConfig holds server-specific configuration settings.
type ServerConfig struct {
	// Address is the address to listen on (e.g., ":8080")
	Address string `yaml:"address"`
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `yaml:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `yaml:"write_timeout"`
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	// Security contains security-related settings
	Security struct {
		// Enabled indicates whether security features are enabled
		Enabled bool `yaml:"enabled"`
		// APIKey is the API key for authentication
		APIKey string `yaml:"api_key"`
		// RateLimit is the maximum number of requests per minute
		RateLimit int `yaml:"rate_limit"`
		// CORS contains CORS configuration
		CORS struct {
			// Enabled indicates whether CORS is enabled
			Enabled bool `yaml:"enabled"`
			// AllowedOrigins is a list of allowed origins
			AllowedOrigins []string `yaml:"allowed_origins"`
			// AllowedMethods is a list of allowed HTTP methods
			AllowedMethods []string `yaml:"allowed_methods"`
			// AllowedHeaders is a list of allowed headers
			AllowedHeaders []string `yaml:"allowed_headers"`
			// MaxAge is the maximum age of preflight requests
			MaxAge int `yaml:"max_age"`
		} `yaml:"cors"`
		// TLS contains TLS configuration
		TLS struct {
			// Enabled indicates whether TLS is enabled
			Enabled bool `yaml:"enabled"`
			// Certificate is the path to the certificate file
			Certificate string `yaml:"certificate"`
			// Key is the path to the private key file
			Key string `yaml:"key"`
		} `yaml:"tls"`
	} `yaml:"security"`
}

// Config represents the complete application configuration.
// It combines all configuration components into a single structure.
type Config struct {
	// App contains application-specific settings
	App AppConfig `yaml:"app"`
	// Log contains logging configuration
	Log LogConfig `yaml:"log"`
	// Elasticsearch contains Elasticsearch connection settings
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	// Server contains HTTP server settings
	Server ServerConfig `yaml:"server"`
	// Crawler contains crawler-specific settings
	Crawler CrawlerConfig `yaml:"crawler"`
	// Sources contains the list of news sources to crawl
	Sources []Source `yaml:"sources"`
	// Command is the command being run (e.g., "httpd", "job")
	Command string `yaml:"-"`
}

// GetCrawlerConfig implements Interface
func (c *Config) GetCrawlerConfig() *CrawlerConfig {
	return &c.Crawler
}

// GetElasticsearchConfig implements Interface
func (c *Config) GetElasticsearchConfig() *ElasticsearchConfig {
	return &c.Elasticsearch
}

// GetLogConfig implements Interface
func (c *Config) GetLogConfig() *LogConfig {
	return &c.Log
}

// GetAppConfig implements Interface
func (c *Config) GetAppConfig() *AppConfig {
	return &c.App
}

// GetSources implements Interface
func (c *Config) GetSources() []Source {
	return c.Sources
}

// GetServerConfig implements Interface
func (c *Config) GetServerConfig() *ServerConfig {
	return &c.Server
}

// GetCommand implements Interface
func (c *Config) GetCommand() string {
	return c.Command
}

// Ensure Config implements Interface
var _ Interface = (*Config)(nil)

// parseRateLimit parses the rate limit duration from a string.
// It converts a string duration (e.g., "1s", "500ms") into a time.Duration.
// If the rate limit string is empty, it returns the default rate limit.
//
// Parameters:
//   - rateLimitStr: The rate limit as a string
//
// Returns:
//   - time.Duration: The parsed duration
//   - error: Any error that occurred during parsing
func parseRateLimit(rateLimitStr string) (time.Duration, error) {
	if rateLimitStr == "" {
		return defaultRateLimit, nil
	}
	return time.ParseDuration(rateLimitStr)
}

// NewHTTPTransport creates a new HTTP transport.
// It returns the default HTTP transport configuration.
//
// Returns:
//   - http.RoundTripper: The configured HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// ParseRateLimit parses a rate limit string and returns a time.Duration.
// If the input is invalid, it returns a default value of 1 second and an error.
//
// Parameters:
//   - rateLimit: The rate limit as a string
//
// Returns:
//   - time.Duration: The parsed duration or default value
//   - error: Any error that occurred during parsing
func ParseRateLimit(rateLimit string) (time.Duration, error) {
	if rateLimit == "" {
		return time.Second, errors.New("rate limit cannot be empty")
	}
	duration, err := time.ParseDuration(rateLimit)
	if err != nil {
		return time.Second, errors.New("error parsing duration")
	}
	return duration, nil
}

// LoadConfig loads configuration from a YAML file.
// It reads the file, parses it as YAML, and returns the configuration.
//
// Parameters:
//   - path: The path to the YAML configuration file
//
// Returns:
//   - *Config: The loaded configuration
//   - error: Any error that occurred during loading
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if unmarshalErr := yaml.Unmarshal(data, &config); unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", unmarshalErr)
	}

	return &config, nil
}

// Validate validates the source configuration.
// It checks that all required fields are present and have valid values.
//
// Returns:
//   - error: Any validation errors that occurred
func (s *Source) Validate() error {
	if s.RateLimit <= 0 {
		return errors.New("rate limit must be positive")
	}
	if s.MaxDepth < 0 {
		return errors.New("max depth must be non-negative")
	}
	return nil
}
