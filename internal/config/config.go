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

// Error definitions for configuration-related errors
var (
	// ErrMissingElasticURL is returned when the Elasticsearch URL is not provided
	ErrMissingElasticURL = errors.New("ELASTIC_URL is required")
)

// Configuration keys for environment variables and configuration values.
// These constants are used throughout the application to maintain consistency
// in configuration key names.
const (
	// AppEnvKey defines the environment type (development, staging, production)
	AppEnvKey = "APP_ENV"
	// LogLevelKey defines the logging level (debug, info, warn, error)
	LogLevelKey = "LOG_LEVEL"
	// AppDebugKey enables/disables debug mode
	AppDebugKey = "APP_DEBUG"
	// CrawlerBaseURLKey defines the starting URL for the crawler
	CrawlerBaseURLKey = "CRAWLER_BASE_URL"
	// CrawlerMaxDepthKey defines how deep the crawler should traverse
	CrawlerMaxDepthKey = "CRAWLER_MAX_DEPTH"
	// CrawlerRateLimitKey defines the delay between requests
	CrawlerRateLimitKey = "CRAWLER_RATE_LIMIT"
	// CrawlerSourceFileKey defines the path to the sources configuration file
	CrawlerSourceFileKey = "CRAWLER_SOURCE_FILE"
	// ElasticURLKey defines the Elasticsearch server URL
	ElasticURLKey = "ELASTIC_URL"
	// ElasticUsernameKey defines the Elasticsearch username for authentication
	ElasticUsernameKey = "ELASTIC_USERNAME"
	// ElasticPasswordKey defines the Elasticsearch password for authentication
	ElasticPasswordKey = "ELASTIC_PASSWORD"
	// ElasticIndexNameKey defines the name of the Elasticsearch index
	ElasticIndexNameKey = "ELASTIC_INDEX_NAME"
	// ElasticSkipTLSKey enables/disables TLS verification for Elasticsearch
	ElasticSkipTLSKey = "ELASTIC_SKIP_TLS"
	// ElasticAPIKeyKey defines the API key for Elasticsearch authentication
	//nolint:gosec // This is a false positive
	ElasticAPIKeyKey = "ELASTIC_API_KEY"
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
	viper.Set(CrawlerMaxDepthKey, depth)
}

// SetRateLimit sets the RateLimit in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - rate: The time duration between requests
func (c *CrawlerConfig) SetRateLimit(rate time.Duration) {
	c.RateLimit = rate
	viper.Set(CrawlerRateLimitKey, rate.String())
}

// SetBaseURL sets the BaseURL in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - url: The base URL for crawling
func (c *CrawlerConfig) SetBaseURL(url string) {
	c.BaseURL = url
	viper.Set(CrawlerBaseURLKey, url)
}

// SetIndexName sets the IndexName in the CrawlerConfig and updates Viper.
// This ensures both the local config and Viper stay in sync.
//
// Parameters:
//   - index: The name of the Elasticsearch index
func (c *CrawlerConfig) SetIndexName(index string) {
	c.IndexName = index
	viper.Set(ElasticIndexNameKey, index)
}

// ElasticsearchConfig holds Elasticsearch-specific configuration settings.
// It defines how to connect to and interact with Elasticsearch.
type ElasticsearchConfig struct {
	// URL is the Elasticsearch server URL
	URL string `yaml:"url"`
	// Username for Elasticsearch authentication
	Username string `yaml:"username"`
	// Password for Elasticsearch authentication
	Password string `yaml:"password"`
	// APIKey for Elasticsearch authentication
	APIKey string `yaml:"api_key"`
	// IndexName is the default index for storing data
	IndexName string `yaml:"index_name"`
	// SkipTLS determines whether to skip TLS verification
	SkipTLS bool `yaml:"skip_tls"`
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
	// ArticleIndex is the Elasticsearch index for storing articles
	ArticleIndex string `yaml:"article_index"`
	// Index is the Elasticsearch index for storing non-article content
	Index string `yaml:"index"`
	// RateLimit defines the delay between requests for this source
	RateLimit time.Duration `yaml:"rate_limit"`
	// MaxDepth defines how deep to crawl this source
	MaxDepth int `yaml:"max_depth"`
	// Time defines which time-related fields to extract
	Time []string `yaml:"time"`
	// Selectors defines how to extract content from pages
	Selectors SourceSelectors `yaml:"selectors"`
}

// SourceSelectors defines the selectors for a source.
// It contains the rules for extracting content from web pages.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
}

// Config represents the complete application configuration.
// It combines all configuration components into a single structure.
type Config struct {
	// App contains application-level settings
	App AppConfig `yaml:"app"`
	// Crawler contains crawler-specific settings
	Crawler CrawlerConfig `yaml:"crawler"`
	// Elasticsearch contains Elasticsearch connection settings
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	// Log contains logging-related settings
	Log LogConfig `yaml:"log"`
	// Sources contains the list of news sources to crawl
	Sources []Source `yaml:"sources"`
}

// parseRateLimit parses the rate limit duration from a string.
// It converts a string duration (e.g., "1s", "500ms") into a time.Duration.
//
// Parameters:
//   - rateLimitStr: The rate limit as a string
//
// Returns:
//   - time.Duration: The parsed duration
//   - error: Any error that occurred during parsing
func parseRateLimit(rateLimitStr string) (time.Duration, error) {
	if rateLimitStr == "" {
		return time.Second, errors.New("rate limit cannot be empty")
	}
	rateLimit, err := time.ParseDuration(rateLimitStr)
	if err != nil {
		return time.Second, errors.New("error parsing duration")
	}
	return rateLimit, nil
}

// ValidateConfig validates the configuration values.
// It checks that all required fields are present and have valid values.
//
// Parameters:
//   - cfg: The configuration to validate
//
// Returns:
//   - error: Any validation errors that occurred
func ValidateConfig(cfg *Config) error {
	if cfg.Elasticsearch.URL == "" {
		return ErrMissingElasticURL
	}
	if cfg.Crawler.Parallelism < 1 {
		return errors.New("crawler parallelism must be positive")
	}
	if cfg.Crawler.MaxDepth < 0 {
		return errors.New("crawler max depth must be non-negative")
	}
	if cfg.Crawler.RateLimit < 0 {
		return errors.New("crawler rate limit must be non-negative")
	}
	if cfg.Crawler.RandomDelay < 0 {
		return errors.New("crawler random delay must be non-negative")
	}
	return nil
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
