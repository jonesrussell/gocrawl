// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Constants for default configuration values
const (
	defaultRateLimit = 2 * time.Second
	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 2
	// DefaultRateLimit is the default rate limit
	DefaultRateLimit = time.Second * 2
	// DefaultParallelism is the default number of parallel requests
	DefaultParallelism = 2
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
	// Sources is the list of configured sources
	Sources []Source `yaml:"sources"`
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
	TLS TLSConfig `yaml:"tls"`
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
	// Time defines which time-related fields to extract
	Time []string `yaml:"time,omitempty"`
	// ArticleIndex is the Elasticsearch index for storing article data
	ArticleIndex string `yaml:"article_index"`
	// Index is the Elasticsearch index for storing content data
	Index string `yaml:"index"`
	// Selectors defines the CSS selectors for content extraction
	Selectors SourceSelectors `yaml:"selectors"`
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
		TLS TLSConfig `yaml:"tls"`
	} `yaml:"security"`
}

// Config represents the complete application configuration.
// It combines all configuration components into a single structure.
type Config struct {
	// App contains application-level configuration
	App AppConfig `yaml:"app"`
	// Log contains logging configuration
	Log LogConfig `yaml:"log"`
	// Elasticsearch contains Elasticsearch configuration
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	// Server contains server configuration
	Server ServerConfig `yaml:"server"`
	// Crawler contains crawler configuration
	Crawler CrawlerConfig `yaml:"crawler"`
	// Sources contains the list of sources
	Sources []Source `yaml:"sources"`
	// Priority contains priority configuration
	Priority PriorityConfig `yaml:"priority"`
	// Command is the current command being run
	Command string
	// logger is the logger instance
	logger Logger
}

// load loads all configuration values from Viper
func (c *Config) load() error {
	// Initialize Viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add standard config paths
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.gocrawl")
	v.AddConfigPath("/etc/gocrawl")

	// Set environment variable prefix
	v.SetEnvPrefix("GOCRAWL")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
		c.logger.Info("No config file found, using defaults")
	}

	// Debug: Print Viper values
	c.logger.Info("Loading config from Viper",
		String("app.environment", v.GetString("app.environment")),
		String("app.name", v.GetString("app.name")),
		String("app.version", v.GetString("app.version")),
		String("app.debug", fmt.Sprintf("%v", v.GetBool("app.debug"))),
		String("log.level", v.GetString("log.level")),
		String("log.debug", fmt.Sprintf("%v", v.GetBool("log.debug"))),
		String("elasticsearch.addresses", fmt.Sprintf("%v", v.GetStringSlice("elasticsearch.addresses"))),
		String("elasticsearch.username", v.GetString("elasticsearch.username")),
		String("elasticsearch.password", v.GetString("elasticsearch.password")),
		String("elasticsearch.api_key", v.GetString("elasticsearch.api_key")),
		String("elasticsearch.index_name", v.GetString("elasticsearch.index_name")),
		String("elasticsearch.cloud.id", v.GetString("elasticsearch.cloud.id")),
		String("elasticsearch.cloud.api_key", v.GetString("elasticsearch.cloud.api_key")),
		String("elasticsearch.tls.enabled", fmt.Sprintf("%v", v.GetBool("elasticsearch.tls.enabled"))),
		String("elasticsearch.tls.certificate", v.GetString("elasticsearch.tls.certificate")),
		String("elasticsearch.tls.key", v.GetString("elasticsearch.tls.key")),
		String("elasticsearch.retry.enabled", fmt.Sprintf("%v", v.GetBool("elasticsearch.retry.enabled"))),
		String("elasticsearch.retry.initial_wait", v.GetString("elasticsearch.retry.initial_wait")),
		String("elasticsearch.retry.max_wait", v.GetString("elasticsearch.retry.max_wait")),
		String("elasticsearch.retry.max_retries", fmt.Sprintf("%d", v.GetInt("elasticsearch.retry.max_retries"))),
		String("crawler.source_file", v.GetString("crawler.source_file")),
	)

	// Load app config
	c.App.Environment = viper.GetString("app.environment")
	c.App.Name = viper.GetString("app.name")
	c.App.Version = viper.GetString("app.version")
	c.App.Debug = viper.GetBool("app.debug")

	// Load log config
	c.Log.Level = viper.GetString("log.level")
	c.Log.Debug = viper.GetBool("log.debug")

	// Load Elasticsearch config
	c.Elasticsearch = *createElasticsearchConfig()

	// Load crawler config
	crawlerConfig, err := createCrawlerConfig()
	if err != nil {
		return fmt.Errorf("failed to create crawler config: %w", err)
	}
	c.Crawler = crawlerConfig

	// Load server config
	c.Server = ServerConfig{
		Address: viper.GetString("server.address"),
		Security: struct {
			Enabled   bool   `yaml:"enabled"`
			APIKey    string `yaml:"api_key"`
			RateLimit int    `yaml:"rate_limit"`
			CORS      struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			} `yaml:"cors"`
			TLS TLSConfig `yaml:"tls"`
		}{
			Enabled:   viper.GetBool("server.security.enabled"),
			APIKey:    viper.GetString("server.security.api_key"),
			RateLimit: viper.GetInt("server.security.rate_limit"),
			CORS: struct {
				Enabled        bool     `yaml:"enabled"`
				AllowedOrigins []string `yaml:"allowed_origins"`
				AllowedMethods []string `yaml:"allowed_methods"`
				AllowedHeaders []string `yaml:"allowed_headers"`
				MaxAge         int      `yaml:"max_age"`
			}{
				Enabled:        viper.GetBool("server.security.cors.enabled"),
				AllowedOrigins: viper.GetStringSlice("server.security.cors.allowed_origins"),
				AllowedMethods: viper.GetStringSlice("server.security.cors.allowed_methods"),
				AllowedHeaders: viper.GetStringSlice("server.security.cors.allowed_headers"),
				MaxAge:         viper.GetInt("server.security.cors.max_age"),
			},
			TLS: TLSConfig{
				Enabled:  viper.GetBool("server.security.tls.enabled"),
				CertFile: viper.GetString("server.security.tls.certificate"),
				KeyFile:  viper.GetString("server.security.tls.key"),
			},
		},
	}

	return nil
}

// GetAppConfig returns the application configuration
func (c *Config) GetAppConfig() *AppConfig {
	return &c.App
}

// GetCrawlerConfig returns the crawler configuration
func (c *Config) GetCrawlerConfig() *CrawlerConfig {
	return &c.Crawler
}

// GetElasticsearchConfig returns the Elasticsearch configuration
func (c *Config) GetElasticsearchConfig() *ElasticsearchConfig {
	return &c.Elasticsearch
}

// GetServerConfig returns the server configuration
func (c *Config) GetServerConfig() *ServerConfig {
	return &c.Server
}

// GetLogger returns the logger
func (c *Config) GetLogger() Logger {
	return c.logger
}

// GetLogConfig implements Interface
func (c *Config) GetLogConfig() *LogConfig {
	return &c.Log
}

// GetSources implements Interface
func (c *Config) GetSources() []Source {
	return c.Sources
}

// GetCommand implements Interface
func (c *Config) GetCommand() string {
	return c.Command
}

// GetPriorityConfig returns the priority configuration
func (c *Config) GetPriorityConfig() *PriorityConfig {
	return &c.Priority
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

// Validate validates a source configuration.
func (s *Source) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}
	if s.URL == "" {
		return errors.New("URL is required")
	}
	if s.RateLimit <= 0 {
		return errors.New("rate limit must be positive")
	}
	if s.MaxDepth <= 0 {
		return errors.New("max depth must be positive")
	}

	// Validate time format if provided
	if len(s.Time) > 0 {
		for _, t := range s.Time {
			if _, timeErr := time.Parse("15:04", t); timeErr != nil {
				return fmt.Errorf("invalid time format: %w", timeErr)
			}
		}
	}

	return nil
}

// NewNoOp creates a no-op config that returns default values.
// This is useful for testing or when configuration is not needed.
func NewNoOp() Interface {
	return &NoOpConfig{}
}

// NoOpConfig implements Interface but returns default values.
type NoOpConfig struct{}

func (c *NoOpConfig) GetAppConfig() *AppConfig {
	return &AppConfig{
		Environment: "test",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       false,
	}
}

func (c *NoOpConfig) GetLogConfig() *LogConfig {
	return &LogConfig{
		Level: "info",
		Debug: false,
	}
}

func (c *NoOpConfig) GetElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "gocrawl",
	}
}

func (c *NoOpConfig) GetServerConfig() *ServerConfig {
	return &ServerConfig{
		Address: ":8080",
	}
}

func (c *NoOpConfig) GetSources() []Source {
	return []Source{}
}

func (c *NoOpConfig) GetCommand() string {
	return "test"
}

func (c *NoOpConfig) GetCrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		MaxDepth:    DefaultMaxDepth,
		RateLimit:   DefaultRateLimit,
		RandomDelay: time.Second,
		Parallelism: DefaultParallelism,
	}
}

func (c *NoOpConfig) GetPriorityConfig() *PriorityConfig {
	return &PriorityConfig{
		Default: 1,
		Rules:   []PriorityRule{},
	}
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Environment: envDevelopment,
			Name:        "gocrawl",
			Version:     "1.0.0",
			Debug:       false,
		},
		Log: LogConfig{
			Level: "info",
			Debug: false,
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			IndexName: "gocrawl",
		},
		Server: ServerConfig{
			Address: ":8080",
		},
	}
}

// NewConfig creates a new Config instance
func NewConfig(logger Logger) (Interface, error) {
	// Create Viper instance
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add standard config paths
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.gocrawl")
	v.AddConfigPath("/etc/gocrawl")

	// Set environment variable prefix and key replacer
	v.SetEnvPrefix("GOCRAWL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
		logger.Info("No config file found, using defaults")
	}

	// Log loaded configuration
	logger.Info("Loading config from Viper", []Field{
		{Key: "app.environment", Value: v.GetString("app.environment")},
		{Key: "app.name", Value: v.GetString("app.name")},
		{Key: "app.version", Value: v.GetString("app.version")},
		{Key: "app.debug", Value: v.GetBool("app.debug")},
		{Key: "log.level", Value: v.GetString("log.level")},
		{Key: "log.debug", Value: v.GetBool("log.debug")},
		{Key: "elasticsearch.addresses", Value: v.GetStringSlice("elasticsearch.addresses")},
		{Key: "elasticsearch.username", Value: v.GetString("elasticsearch.username")},
		{Key: "elasticsearch.password", Value: v.GetString("elasticsearch.password")},
		{Key: "elasticsearch.api_key", Value: v.GetString("elasticsearch.api_key")},
		{Key: "elasticsearch.index_name", Value: v.GetString("elasticsearch.index_name")},
		{Key: "elasticsearch.cloud.id", Value: v.GetString("elasticsearch.cloud.id")},
		{Key: "elasticsearch.cloud.api_key", Value: v.GetString("elasticsearch.cloud.api_key")},
		{Key: "elasticsearch.tls.enabled", Value: v.GetBool("elasticsearch.tls.enabled")},
		{Key: "elasticsearch.tls.certificate", Value: v.GetString("elasticsearch.tls.certificate")},
		{Key: "elasticsearch.tls.key", Value: v.GetString("elasticsearch.tls.key")},
		{Key: "elasticsearch.retry.enabled", Value: v.GetBool("elasticsearch.retry.enabled")},
		{Key: "elasticsearch.retry.initial_wait", Value: v.GetString("elasticsearch.retry.initial_wait")},
		{Key: "elasticsearch.retry.max_wait", Value: v.GetString("elasticsearch.retry.max_wait")},
		{Key: "elasticsearch.retry.max_retries", Value: v.GetInt("elasticsearch.retry.max_retries")},
		{Key: "crawler.source_file", Value: v.GetString("crawler.source_file")},
	}...)

	// Create config instance
	cfg := DefaultConfig()

	// Load values from Viper
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure Elasticsearch config is properly loaded
	cfg.Elasticsearch = *createElasticsearchConfig()

	// Ensure Crawler config is properly loaded
	crawlerCfg, err := createCrawlerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create crawler config: %w", err)
	}
	cfg.Crawler = crawlerCfg

	// Validate config
	if err := ValidateConfig(cfg); err != nil {
		return nil, err // Return validation error directly without wrapping
	}

	return cfg, nil
}

// Params holds the parameters for creating a config.
type Params struct {
	// Environment specifies the runtime environment (development, staging, production)
	Environment string
	// Debug enables debug mode for additional logging
	Debug bool
	// Command is the command being run (e.g., "httpd", "job")
	Command string
}
