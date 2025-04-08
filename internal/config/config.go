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
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
}

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

// Validate checks if the log configuration is valid
func (c *LogConfig) Validate() error {
	if c == nil {
		return errors.New("log configuration is required")
	}

	if c.Level == "" {
		return errors.New("log level cannot be empty")
	}

	switch strings.ToLower(c.Level) {
	case "debug", "info", "warn", "error":
		// Valid level
	default:
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	return nil
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
	// AllowedDomains is a list of allowed domains for this source
	AllowedDomains []string `yaml:"allowed_domains"`
	// StartURLs is a list of starting URLs for this source
	StartURLs []string `yaml:"start_urls"`
	// Rules defines the crawling rules for this source
	Rules Rules `yaml:"rules"`
	// Headers is a map of headers to include in requests for this source
	Headers map[string]string `yaml:"headers"`
}

// SourceSelectors defines the selectors for a source.
// It contains the rules for extracting content from web pages.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
	// Content contains selectors for general content
	Content ContentSelectors `yaml:"content"`
}

// ArticleSelectors defines the selectors for article content.
type ArticleSelectors struct {
	// Container is the selector for the article container
	Container string `yaml:"container"`
	// Title is the selector for the article title
	Title string `yaml:"title"`
	// Author is the selector for the article author
	Author string `yaml:"author"`
	// Body is the selector for the article body
	Body string `yaml:"body"`
	// Intro is the selector for the article introduction
	Intro string `yaml:"intro"`
	// Byline is the selector for the article byline
	Byline string `yaml:"byline"`
	// PublishedTime is the selector for the article published time
	PublishedTime string `yaml:"published_time"`
	// TimeAgo is the selector for the relative time
	TimeAgo string `yaml:"time_ago"`
	// JSONLD is the selector for JSON-LD metadata
	JSONLD string `yaml:"json_ld"`
	// Description is the selector for the article description
	Description string `yaml:"description"`
	// Section is the selector for the article section
	Section string `yaml:"section"`
	// Keywords is the selector for article keywords
	Keywords string `yaml:"keywords"`
}

// ContentSelectors defines the selectors for general content.
type ContentSelectors struct {
	// Title is the selector for the content title
	Title string `yaml:"title"`
	// Description is the selector for the content description
	Description string `yaml:"description"`
	// URL is the selector for the content URL
	URL string `yaml:"url"`
}

// Rule defines a crawling rule for a source.
type Rule struct {
	// Pattern is the URL pattern to match
	Pattern string `yaml:"pattern"`
	// Action is the action to take when the pattern matches
	Action string `yaml:"action"`
	// Priority is the priority of the rule
	Priority int `yaml:"priority"`
}

// Rules is a slice of Rule
type Rules []Rule

// Validate checks if the rules are valid.
func (r Rules) Validate() error {
	if len(r) == 0 {
		return nil
	}

	for _, rule := range r {
		if rule.Pattern == "" {
			return errors.New("rule pattern is required")
		}
		if rule.Action == "" {
			return errors.New("rule action is required")
		}
		if rule.Priority < 0 {
			return errors.New("rule priority must be non-negative")
		}
	}

	return nil
}

// TLSConfig holds TLS configuration settings.
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
type Config struct {
	App           *AppConfig
	Log           *LogConfig
	Elasticsearch *ElasticsearchConfig
	Server        *ServerConfig
	Crawler       *CrawlerConfig
	Sources       []Source
	Priority      *PriorityConfig
	Command       string
	logger        Logger
}

// New creates a new Config instance with the provided logger.
func New(logger Logger) (*Config, error) {
	if logger == nil {
		logger = &defaultLogger{}
	}

	cfg := &Config{
		logger: logger,
	}

	if err := cfg.load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// load loads the configuration from Viper.
func (c *Config) load() error {
	// Initialize Viper if no config file is set
	if viper.GetViper().ConfigFileUsed() == "" {
		viper.AutomaticEnv()
		viper.SetEnvPrefix("GOCRAWL")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	// Set environment from env var or default to development
	env := viper.GetString("app.environment")
	if env == "" {
		env = os.Getenv("GOCRAWL_APP_ENVIRONMENT")
		if env == "" {
			env = "development"
		}
		viper.Set("app.environment", env)
	}

	// Load app config
	appCfg, err := loadAppConfig()
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}
	c.App = appCfg

	// Load log config
	logCfg, err := loadLogConfig()
	if err != nil {
		return fmt.Errorf("failed to load log config: %w", err)
	}
	c.Log = logCfg

	// Load elasticsearch config
	esCfg, err := loadElasticsearchConfig()
	if err != nil {
		return fmt.Errorf("failed to load elasticsearch config: %w", err)
	}
	c.Elasticsearch = esCfg

	// Load server config
	serverCfg, err := loadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to load server config: %w", err)
	}
	c.Server = serverCfg

	// Load crawler config
	crawlerCfg, err := loadCrawlerConfig()
	if err != nil {
		return fmt.Errorf("failed to load crawler config: %w", err)
	}
	c.Crawler = crawlerCfg

	// Load sources
	sources, err := loadSources()
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}
	c.Sources = sources

	// Load priority config
	priorityCfg, err := loadPriorityConfig()
	if err != nil {
		return fmt.Errorf("failed to load priority config: %w", err)
	}
	c.Priority = priorityCfg

	return nil
}

// validate validates the configuration.
func (c *Config) validate() error {
	if c == nil {
		return errors.New("config is required")
	}

	if c.Log == nil {
		return errors.New("log configuration is required")
	}

	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("invalid log configuration: %w", err)
	}

	return nil
}

// GetAppConfig returns the application configuration.
func (c *Config) GetAppConfig() *AppConfig {
	return c.App
}

// GetLogConfig returns the logging configuration.
func (c *Config) GetLogConfig() *LogConfig {
	return c.Log
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (c *Config) GetElasticsearchConfig() *ElasticsearchConfig {
	return c.Elasticsearch
}

// GetServerConfig returns the server configuration.
func (c *Config) GetServerConfig() *ServerConfig {
	return c.Server
}

// GetCrawlerConfig returns the crawler configuration.
func (c *Config) GetCrawlerConfig() *CrawlerConfig {
	return c.Crawler
}

// GetSources returns the list of sources.
func (c *Config) GetSources() []Source {
	return c.Sources
}

// GetCommand returns the current command.
func (c *Config) GetCommand() string {
	return c.Command
}

// GetPriorityConfig returns the priority configuration.
func (c *Config) GetPriorityConfig() *PriorityConfig {
	return c.Priority
}

// GetLogger returns the logger.
func (c *Config) GetLogger() Logger {
	return c.logger
}

// loadAppConfig loads the application configuration from Viper.
func loadAppConfig() (*AppConfig, error) {
	appCfg := &AppConfig{}

	appCfg.Environment = viper.GetString("app.environment")
	appCfg.Name = viper.GetString("app.name")
	appCfg.Version = viper.GetString("app.version")
	appCfg.Debug = viper.GetBool("app.debug")

	return appCfg, nil
}

// loadLogConfig loads the logging configuration from Viper.
func loadLogConfig() (*LogConfig, error) {
	logCfg := &LogConfig{}

	logCfg.Level = viper.GetString("log.level")
	logCfg.Debug = viper.GetBool("log.debug")

	return logCfg, nil
}

// loadElasticsearchConfig loads the Elasticsearch configuration from Viper.
func loadElasticsearchConfig() (*ElasticsearchConfig, error) {
	esCfg := &ElasticsearchConfig{}

	esCfg.Addresses = viper.GetStringSlice("elasticsearch.addresses")
	esCfg.Username = viper.GetString("elasticsearch.username")
	esCfg.Password = viper.GetString("elasticsearch.password")
	esCfg.APIKey = viper.GetString("elasticsearch.api_key")
	esCfg.IndexName = viper.GetString("elasticsearch.index_name")
	esCfg.Cloud.ID = viper.GetString("elasticsearch.cloud.id")
	esCfg.Cloud.APIKey = viper.GetString("elasticsearch.cloud.api_key")
	esCfg.TLS.Enabled = viper.GetBool("elasticsearch.tls.enabled")
	esCfg.TLS.CertFile = viper.GetString("elasticsearch.tls.certificate")
	esCfg.TLS.KeyFile = viper.GetString("elasticsearch.tls.key")
	esCfg.Retry.Enabled = viper.GetBool("elasticsearch.retry.enabled")
	esCfg.Retry.InitialWait = viper.GetDuration("elasticsearch.retry.initial_wait")
	esCfg.Retry.MaxWait = viper.GetDuration("elasticsearch.retry.max_wait")
	esCfg.Retry.MaxRetries = viper.GetInt("elasticsearch.retry.max_retries")

	return esCfg, nil
}

// loadServerConfig loads the server configuration from Viper.
func loadServerConfig() (*ServerConfig, error) {
	serverCfg := &ServerConfig{}

	serverCfg.Address = viper.GetString("server.address")
	serverCfg.ReadTimeout = viper.GetDuration("server.read_timeout")
	serverCfg.WriteTimeout = viper.GetDuration("server.write_timeout")
	serverCfg.IdleTimeout = viper.GetDuration("server.idle_timeout")
	serverCfg.Security.Enabled = viper.GetBool("server.security.enabled")
	serverCfg.Security.APIKey = viper.GetString("server.security.api_key")
	serverCfg.Security.RateLimit = viper.GetInt("server.security.rate_limit")
	serverCfg.Security.CORS.Enabled = viper.GetBool("server.security.cors.enabled")
	serverCfg.Security.CORS.AllowedOrigins = viper.GetStringSlice("server.security.cors.allowed_origins")
	serverCfg.Security.CORS.AllowedMethods = viper.GetStringSlice("server.security.cors.allowed_methods")
	serverCfg.Security.CORS.AllowedHeaders = viper.GetStringSlice("server.security.cors.allowed_headers")
	serverCfg.Security.CORS.MaxAge = viper.GetInt("server.security.cors.max_age")
	serverCfg.Security.TLS.Enabled = viper.GetBool("server.security.tls.enabled")
	serverCfg.Security.TLS.CertFile = viper.GetString("server.security.tls.certificate")
	serverCfg.Security.TLS.KeyFile = viper.GetString("server.security.tls.key")

	return serverCfg, nil
}

// loadCrawlerConfig loads the crawler configuration from Viper.
func loadCrawlerConfig() (*CrawlerConfig, error) {
	crawlerCfg := &CrawlerConfig{}

	crawlerCfg.BaseURL = viper.GetString("crawler.base_url")
	crawlerCfg.MaxDepth = viper.GetInt("crawler.max_depth")
	crawlerCfg.RateLimit = viper.GetDuration("crawler.rate_limit")
	crawlerCfg.RandomDelay = viper.GetDuration("crawler.random_delay")
	crawlerCfg.IndexName = viper.GetString("elasticsearch.index_name")
	crawlerCfg.ContentIndexName = viper.GetString("elasticsearch.content_index_name")
	crawlerCfg.SourceFile = viper.GetString("crawler.source_file")
	crawlerCfg.Parallelism = viper.GetInt("crawler.parallelism")
	crawlerCfg.Sources = viper.Get("crawler.sources").([]interface{})

	return crawlerCfg, nil
}

// loadSources loads the sources configuration from Viper.
func loadSources() ([]Source, error) {
	sources := []Source{}

	for _, source := range viper.Get("sources").([]interface{}) {
		sourceMap := source.(map[string]interface{})
		sources = append(sources, Source{
			Name:           sourceMap["name"].(string),
			URL:            sourceMap["url"].(string),
			RateLimit:      sourceMap["rate_limit"].(time.Duration),
			MaxDepth:       sourceMap["max_depth"].(int),
			Time:           sourceMap["time"].([]string),
			ArticleIndex:   sourceMap["article_index"].(string),
			Index:          sourceMap["index"].(string),
			Selectors:      sourceMap["selectors"].(SourceSelectors),
			AllowedDomains: sourceMap["allowed_domains"].([]string),
			StartURLs:      sourceMap["start_urls"].([]string),
			Rules:          sourceMap["rules"].(Rules),
			Headers:        sourceMap["headers"].(map[string]string),
		})
	}

	return sources, nil
}

// loadPriorityConfig loads the priority configuration from Viper.
func loadPriorityConfig() (*PriorityConfig, error) {
	priorityCfg := &PriorityConfig{}

	priorityCfg.Default = viper.GetInt("priority.default")
	priorityCfg.Rules = viper.Get("priority.rules").([]PriorityRule)

	return priorityCfg, nil
}

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
	// Initialize Viper
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("GOCRAWL")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Load sources from file
	if config.Crawler.SourceFile != "" {
		sources, err := loadSources(config.Crawler.SourceFile)
		if err != nil {
			return nil, fmt.Errorf("error loading sources: %w", err)
		}
		config.Sources = sources
	}

	// Validate the loaded configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}

	// Check if log section exists
	if !viper.IsSet("log") {
		return errors.New("log configuration is required")
	}

	// Validate log configuration
	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("invalid log configuration: %w", err)
	}

	return nil
}

// DefaultArticleSelectors returns a new ArticleSelectors with default values.
func DefaultArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		Container:     "div.details",
		Title:         "h1.details-title",
		Body:          "div.details-body",
		Intro:         "div.details-intro",
		Byline:        "div.details-byline",
		Author:        "span.author",
		PublishedTime: "time.timeago",
		TimeAgo:       "time.timeago",
		JSONLD:        "script[type='application/ld+json']",
		Description:   "meta[name='description']",
		Section:       "meta[property='article:section']",
		Keywords:      "meta[name='keywords']",
	}
}

// NewConfig creates a new configuration instance.
func NewConfig(logger Logger) (Interface, error) {
	if logger == nil {
		logger = &defaultLogger{}
	}

	cfg := &Config{
		logger: logger,
	}

	if err := cfg.load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the article selectors are valid.
func (s *ArticleSelectors) Validate() error {
	if s == nil {
		return errors.New("article selectors are required")
	}

	if s.Title == "" {
		return errors.New("title selector is required")
	}

	if s.Body == "" {
		return errors.New("body selector is required")
	}

	return nil
}
