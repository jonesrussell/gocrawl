// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
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
	Elasticsearch *elasticsearch.Config `yaml:"elasticsearch"`
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
	// Rules define crawling rules for this source
	Rules Rules `yaml:"rules"`
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

// DefaultArticleSelectors returns a default set of article selectors.
func DefaultArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		Container:     "article, .article, [itemtype*='Article']",
		Title:         "h1",
		Body:          "article, [role='main'], .content, .article-content",
		Intro:         ".article-intro, .post-intro, .entry-summary",
		Byline:        ".article-byline, .post-meta, .entry-meta",
		PublishedTime: "meta[property='article:published_time']",
		TimeAgo:       "time",
		JSONLD:        "script[type='application/ld+json']",
		Section:       "meta[property='article:section']",
		Keywords:      "meta[name='keywords']",
		Description:   "meta[name='description']",
		OGTitle:       "meta[property='og:title']",
		OGDescription: "meta[property='og:description']",
		OGImage:       "meta[property='og:image']",
		OgURL:         "meta[property='og:url']",
		Canonical:     "link[rel='canonical']",
		WordCount:     "meta[property='article:word_count']",
		PublishDate:   "meta[property='article:published_time']",
		Category:      "meta[property='article:section']",
		Tags:          "meta[property='article:tag']",
		Author:        "meta[name='author']",
		BylineName:    ".author-name, .byline-name",
	}
}

// NewConfig creates a new configuration with default values and validates it.
func NewConfig() (*Config, error) {
	cfg := &Config{
		Environment: "development",
		Logger:      log.NewConfig(),
		Server:      server.NewConfig(),
		Priority:    priority.NewConfig(),
		Storage:     storage.NewConfig(),
		Crawler:     NewCrawlerConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
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
		RandomDelay: time.Second / 2,
		Parallelism: 2,
		SourceFile:  "config/sources.yml",
		IndexName:   "crawl_content",
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
func (c *Config) GetElasticsearchConfig() *elasticsearch.Config {
	return c.Elasticsearch
}

// LogConfig represents logging configuration settings.
type LogConfig struct {
	// Level is the logging level (debug, info, warn, error)
	Level string `yaml:"level"`
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

	return nil
}

// LoadConfig loads the configuration from the specified file path.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// New creates a new Config instance with default values.
func New() *Config {
	return &Config{
		Environment:   "development",
		Logger:        log.NewConfig(),
		Server:        server.NewConfig(),
		Priority:      priority.NewConfig(),
		Storage:       storage.NewConfig(),
		Crawler:       NewCrawlerConfig(),
		App:           app.NewConfig(),
		Log:           &LogConfig{Level: "info"},
		Elasticsearch: elasticsearch.NewConfig(),
	}
}

// Validate checks if the source configuration is valid.
func (s *Source) Validate() error {
	if s.Name == "" {
		return errors.New("source name is required")
	}
	if s.URL == "" {
		return errors.New("source URL is required")
	}
	if len(s.AllowedDomains) == 0 {
		return errors.New("at least one allowed domain is required")
	}
	if len(s.StartURLs) == 0 {
		return errors.New("at least one start URL is required")
	}
	if s.RateLimit <= 0 {
		return errors.New("rate limit must be positive")
	}
	if s.MaxDepth < 0 {
		return errors.New("max depth cannot be negative")
	}
	if s.ArticleIndex == "" {
		return errors.New("article index is required")
	}
	if s.Index == "" {
		return errors.New("index is required")
	}
	if err := s.Selectors.Article.Validate(); err != nil {
		return fmt.Errorf("invalid article selectors: %w", err)
	}
	if err := s.Rules.Validate(); err != nil {
		return fmt.Errorf("invalid rules: %w", err)
	}
	return nil
}

// NewNoOp creates a new no-op implementation of the config interface.
func NewNoOp() Interface {
	return &noOpConfig{}
}

// noOpConfig is a no-op implementation of the config interface.
type noOpConfig struct{}

// GetAppConfig returns a no-op app config.
func (c *noOpConfig) GetAppConfig() *app.Config {
	return app.NewConfig()
}

// GetLogConfig returns a no-op log config.
func (c *noOpConfig) GetLogConfig() *log.Config {
	return log.NewConfig()
}

// GetServerConfig returns a no-op server config.
func (c *noOpConfig) GetServerConfig() *server.Config {
	return server.NewConfig()
}

// GetPriorityConfig returns a no-op priority config.
func (c *noOpConfig) GetPriorityConfig() *priority.Config {
	return priority.NewConfig()
}

// GetStorageConfig returns a no-op storage config.
func (c *noOpConfig) GetStorageConfig() *storage.Config {
	return storage.NewConfig()
}

// GetCrawlerConfig returns a no-op crawler config.
func (c *noOpConfig) GetCrawlerConfig() *CrawlerConfig {
	return NewCrawlerConfig()
}

// GetElasticsearchConfig returns a no-op Elasticsearch config.
func (c *noOpConfig) GetElasticsearchConfig() *elasticsearch.Config {
	return elasticsearch.NewConfig()
}

// GetSources returns an empty slice of sources.
func (c *noOpConfig) GetSources() []Source {
	return nil
}

// GetCommand returns an empty string for the command.
func (c *noOpConfig) GetCommand() string {
	return ""
}

// Validate performs no validation.
func (c *noOpConfig) Validate() error {
	return nil
}
