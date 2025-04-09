// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
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

// Config represents the application configuration.
type Config struct {
	// Environment is the application environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Logger holds logging-specific configuration
	Logger *log.Config `yaml:"logger"`
	// Server holds server-specific configuration
	Server *server.Config `yaml:"server"`
	// Priority holds priority-specific configuration
	Priority *priority.Config `yaml:"priority"`
	// Storage holds storage-specific configuration
	Storage *storage.Config `yaml:"storage"`
	// Crawler holds crawler-specific configuration
	Crawler *CrawlerConfig `yaml:"crawler"`
	// App holds application-specific configuration
	App *app.Config `yaml:"app"`
	// Log holds logging configuration
	Log *LogConfig `yaml:"log"`
	// Elasticsearch holds Elasticsearch configuration
	Elasticsearch *ElasticsearchConfig `yaml:"elasticsearch"`
	// Sources holds the list of sources to crawl
	Sources []Source `yaml:"sources"`
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
	// Parallelism defines the number of concurrent crawlers
	Parallelism int `yaml:"parallelism"`
	// SourceFile is the path to the sources configuration file
	SourceFile string `yaml:"source_file"`
	// Sources contains the list of sources to crawl
	Sources []Source `yaml:"sources"`
	// IndexName is the name of the Elasticsearch index
	IndexName string `yaml:"index_name"`
}

// Validate checks if the crawler configuration is valid.
func (c *CrawlerConfig) Validate() error {
	if c == nil {
		return errors.New("crawler configuration is required")
	}

	if c.MaxDepth < 1 {
		return fmt.Errorf("max depth must be greater than 0, got %d", c.MaxDepth)
	}

	if c.RateLimit < 0 {
		return fmt.Errorf("rate limit must be non-negative, got %v", c.RateLimit)
	}

	if c.RandomDelay < 0 {
		return fmt.Errorf("random delay must be non-negative, got %v", c.RandomDelay)
	}

	if c.Parallelism < 1 {
		return fmt.Errorf("parallelism must be greater than 0, got %d", c.Parallelism)
	}

	if c.SourceFile == "" {
		return errors.New("source file is required")
	}

	return nil
}

// Source represents a single source to crawl.
type Source struct {
	// Name is the unique identifier for the source
	Name string `yaml:"name"`
	// URL is the base URL for the source
	URL string `yaml:"url"`
	// AllowedDomains specifies which domains are allowed to be crawled
	AllowedDomains []string `yaml:"allowed_domains"`
	// StartURLs are the initial URLs to start crawling from
	StartURLs []string `yaml:"start_urls"`
	// RateLimit defines the delay between requests for this source
	RateLimit time.Duration `yaml:"rate_limit"`
	// MaxDepth defines how many levels deep to crawl for this source
	MaxDepth int `yaml:"max_depth"`
	// Time holds time-related configuration
	Time []string `yaml:"time"`
	// ArticleIndex is the name of the index for articles
	ArticleIndex string `yaml:"article_index"`
	// Index is the name of the index for general content
	Index string `yaml:"index"`
	// Selectors define CSS selectors for extracting content
	Selectors SourceSelectors `yaml:"selectors"`
}

// NewConfig creates a new configuration with default values.
func NewConfig() *Config {
	return &Config{
		Environment: "development",
		Logger:      log.New(),
		Server:      server.New(),
		Priority:    priority.New(),
		Storage:     storage.New(),
		Crawler:     NewCrawlerConfig(),
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("configuration is required")
	}

	if c.Environment == "" {
		return errors.New("environment is required")
	}

	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvs[c.Environment] {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("invalid logger configuration: %w", err)
	}

	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	if err := c.Priority.Validate(); err != nil {
		return fmt.Errorf("invalid priority configuration: %w", err)
	}

	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("invalid storage configuration: %w", err)
	}

	if err := c.Crawler.Validate(); err != nil {
		return fmt.Errorf("invalid crawler configuration: %w", err)
	}

	return nil
}

// NewCrawlerConfig creates a new crawler configuration with default values.
func NewCrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		MaxDepth:    3,
		RateLimit:   time.Second,
		RandomDelay: time.Millisecond * 500,
		Parallelism: 1,
	}
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

// GetAppConfig returns the application configuration.
func (c *Config) GetAppConfig() *app.Config {
	return c.App
}

// GetLogConfig returns the logging configuration.
func (c *Config) GetLogConfig() *log.Config {
	return c.Logger
}

// GetServerConfig returns the server configuration.
func (c *Config) GetServerConfig() *server.Config {
	return c.Server
}

// GetSources returns the list of sources.
func (c *Config) GetSources() []Source {
	return c.Sources
}

// GetCrawlerConfig returns the crawler configuration.
func (c *Config) GetCrawlerConfig() *CrawlerConfig {
	return c.Crawler
}

// GetPriorityConfig returns the priority configuration.
func (c *Config) GetPriorityConfig() *priority.Config {
	return c.Priority
}

// GetCommand returns the current command.
func (c *Config) GetCommand() string {
	return viper.GetString("command")
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (c *Config) GetElasticsearchConfig() *ElasticsearchConfig {
	return c.Elasticsearch
}

// ElasticsearchConfig holds Elasticsearch-specific configuration settings.
type ElasticsearchConfig struct {
	// Addresses are the Elasticsearch server addresses
	Addresses []string `yaml:"addresses"`
	// IndexName is the name of the Elasticsearch index
	IndexName string `yaml:"index_name"`
	// APIKey is the API key for authentication
	APIKey string `yaml:"api_key"`
	// Username is the username for basic authentication
	Username string `yaml:"username"`
	// Password is the password for basic authentication
	Password string `yaml:"password"`
	// Cloud contains cloud-specific configuration
	Cloud struct {
		// ID is the cloud deployment ID
		ID string `yaml:"id"`
		// APIKey is the cloud API key
		APIKey string `yaml:"api_key"`
	} `yaml:"cloud"`
	// TLS contains TLS configuration
	TLS TLSConfig `yaml:"tls"`
	// Retry contains retry configuration
	Retry struct {
		// Enabled indicates whether retries are enabled
		Enabled bool `yaml:"enabled"`
		// InitialWait is the initial wait time between retries
		InitialWait time.Duration `yaml:"initial_wait"`
		// MaxWait is the maximum wait time between retries
		MaxWait time.Duration `yaml:"max_wait"`
		// MaxRetries is the maximum number of retries
		MaxRetries int `yaml:"max_retries"`
	} `yaml:"retry"`
}

// LogConfig holds logging-specific configuration settings.
type LogConfig struct {
	// Level is the logging level (debug, info, warn, error)
	Level string `yaml:"level"`
	// Format is the log format (json, text)
	Format string `yaml:"format"`
	// Output is the log output destination (stdout, stderr, file)
	Output string `yaml:"output"`
	// File is the log file path (only used when output is file)
	File string `yaml:"file"`
	// MaxSize is the maximum size of the log file in megabytes
	MaxSize int `yaml:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `yaml:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `yaml:"max_age"`
	// Compress determines if the rotated log files should be compressed
	Compress bool `yaml:"compress"`
	// Debug enables debug mode for additional logging
	Debug bool `yaml:"debug"`
}

// Validate checks if the log configuration is valid.
func (c *LogConfig) Validate() error {
	if c == nil {
		return errors.New("log configuration is required")
	}

	if c.Level == "" {
		return errors.New("log level is required")
	}

	if !ValidLogLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Format == "" {
		return errors.New("log format is required")
	}

	if c.Output == "" {
		return errors.New("log output is required")
	}

	if c.Output == "file" && c.File == "" {
		return errors.New("log file is required when output is file")
	}

	if c.MaxSize < 0 {
		return fmt.Errorf("max size must be non-negative, got %d", c.MaxSize)
	}

	if c.MaxBackups < 0 {
		return fmt.Errorf("max backups must be non-negative, got %d", c.MaxBackups)
	}

	if c.MaxAge < 0 {
		return fmt.Errorf("max age must be non-negative, got %d", c.MaxAge)
	}

	return nil
}

// SourceSelectors defines the selectors for a source.
// It contains the rules for extracting content from web pages.
type SourceSelectors struct {
	// Article contains selectors for article-specific content
	Article ArticleSelectors `yaml:"article"`
}

// ArticleSelectors defines the selectors for article content.
type ArticleSelectors struct {
	// Container is the selector for the article container
	Container string `yaml:"container"`
	// Title is the selector for the article title
	Title string `yaml:"title"`
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
	// OGTitle is the selector for the Open Graph title
	OGTitle string `yaml:"og_title"`
	// OGDescription is the selector for the Open Graph description
	OGDescription string `yaml:"og_description"`
	// OGImage is the selector for the Open Graph image
	OGImage string `yaml:"og_image"`
	// OgURL is the selector for the Open Graph URL
	OgURL string `yaml:"og_url"`
	// Canonical is the selector for the canonical URL
	Canonical string `yaml:"canonical"`
	// WordCount is the selector for the word count
	WordCount string `yaml:"word_count"`
	// PublishDate is the selector for the publish date
	PublishDate string `yaml:"publish_date"`
	// Category is the selector for the article category
	Category string `yaml:"category"`
	// Tags is the selector for article tags
	Tags string `yaml:"tags"`
	// Author is the selector for the article author
	Author string `yaml:"author"`
	// BylineName is the selector for the byline name
	BylineName string `yaml:"byline_name"`
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

// Command represents a command for the application.
type Command struct {
	// Name is the command name
	Name string `yaml:"name"`
	// Description is the command description
	Description string `yaml:"description"`
	// Run is the function to run the command
	Run func() error
}

// Logger defines the interface for logging operations.
type Logger interface {
	// Debug logs a message at debug level
	Debug(msg string, fields ...Field)
	// Info logs a message at info level
	Info(msg string, fields ...Field)
	// Warn logs a message at warn level
	Warn(msg string, fields ...Field)
	// Error logs a message at error level
	Error(msg string, fields ...Field)
	// With returns a new logger with the given fields
	With(fields ...Field) Logger
}

// Field represents a single logging field.
type Field struct {
	// Key is the field name
	Key string
	// Value is the field value
	Value interface{}
}

// New creates a new Config instance with the provided logger.
func New(logger Logger) (*Config, error) {
	if logger == nil {
		logger = &defaultLogger{}
	}

	cfg := &Config{
		Logger:   log.New(),
		Server:   server.New(),
		Priority: priority.New(),
		Storage:  storage.New(),
		Crawler:  NewCrawlerConfig(),
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

// SetElasticsearchIndex sets the Elasticsearch index name.
func (c *Config) SetElasticsearchIndex(index string) {
	c.Elasticsearch.IndexName = index
	viper.Set("elasticsearch.index_name", index)
}

// GetElasticsearchIndex returns the Elasticsearch index name.
func (c *Config) GetElasticsearchIndex() string {
	return c.Elasticsearch.IndexName
}

// NoOpConfig is a no-op implementation of the Config interface.
type NoOpConfig struct{}

// GetAppConfig returns a default app configuration.
func (c *NoOpConfig) GetAppConfig() *app.Config {
	return &app.Config{
		Environment: "test",
		Name:        "test",
		Version:     "test",
		Debug:       false,
	}
}

// GetLogConfig returns a default log configuration.
func (c *NoOpConfig) GetLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		File:       "",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	}
}

// GetCrawlerConfig returns a default crawler configuration.
func (c *NoOpConfig) GetCrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		MaxDepth:    DefaultMaxDepth,
		RateLimit:   DefaultRateLimit,
		RandomDelay: time.Second,
		Parallelism: DefaultParallelism,
	}
}

// GetElasticsearchConfig returns a default Elasticsearch configuration.
func (c *NoOpConfig) GetElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "gocrawl",
	}
}

// GetServerConfig returns a default server configuration.
func (c *NoOpConfig) GetServerConfig() *server.Config {
	return &server.Config{
		SecurityEnabled: false,
		APIKey:          "",
	}
}

// GetSources returns an empty list of sources.
func (c *NoOpConfig) GetSources() []Source {
	return []Source{}
}

// GetPriorityConfig returns a default priority configuration.
func (c *NoOpConfig) GetPriorityConfig() *priority.Config {
	return priority.New()
}

// GetEnvironment returns the test environment.
func (c *NoOpConfig) GetEnvironment() string {
	return "test"
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

// defaultLogger is a no-op implementation of the Logger interface.
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, fields ...Field) {}
func (l *defaultLogger) Info(msg string, fields ...Field)  {}
func (l *defaultLogger) Warn(msg string, fields ...Field)  {}
func (l *defaultLogger) Error(msg string, fields ...Field) {}
func (l *defaultLogger) With(fields ...Field) Logger       { return l }

// loadAppConfig loads the application configuration.
func loadAppConfig() (*app.Config, error) {
	return &app.Config{
		Environment: "development",
		Name:        "gocrawl",
		Version:     "1.0.0",
		Debug:       false,
	}, nil
}

// loadLogConfig loads the logging configuration.
func loadLogConfig() (*LogConfig, error) {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		File:       "",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	}, nil
}

// loadElasticsearchConfig loads the Elasticsearch configuration.
func loadElasticsearchConfig() (*ElasticsearchConfig, error) {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		IndexName: "gocrawl",
	}, nil
}

// loadCrawlerConfig loads the crawler configuration.
func loadCrawlerConfig() (*CrawlerConfig, error) {
	return &CrawlerConfig{
		MaxDepth:    3,
		RateLimit:   time.Second,
		RandomDelay: time.Millisecond * 500,
		Parallelism: 1,
	}, nil
}

// loadSources loads the sources configuration.
func loadSources() ([]Source, error) {
	return []Source{}, nil
}

// loadPriorityConfig loads the priority configuration.
func loadPriorityConfig() (*priority.Config, error) {
	return priority.New(), nil
}
