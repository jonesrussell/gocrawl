// Package config provides configuration functionality for the application.
package config

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/crawler"
)

const (
	// DefaultMaxDepth is the default maximum depth for crawling
	DefaultMaxDepth = 2
	// DefaultRateLimit is the default rate limit
	DefaultRateLimit = time.Second * 2
	// DefaultParallelism is the default number of parallel requests
	DefaultParallelism = 2
)

// Interface defines the interface for configuration operations.
// It provides access to application configuration settings and
// supports different configuration sources.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *AppConfig
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *LogConfig
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *ElasticsearchConfig
	// GetServerConfig returns the server configuration.
	GetServerConfig() *ServerConfig
	// GetSources returns the list of configured sources.
	GetSources() ([]Source, error)
	// GetCommand returns the current command being executed.
	GetCommand() string
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *CrawlerConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Environment string
	Name        string
	Version     string
	Debug       bool
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
	Debug bool
}

// ElasticsearchConfig holds Elasticsearch configuration.
type ElasticsearchConfig struct {
	Addresses []string
	IndexName string
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Address string
}

// Source holds configuration for a content source.
type Source struct {
	Name      string
	URL       string
	RateLimit time.Duration
	MaxDepth  int
	Time      []string
}

// CrawlerConfig holds crawler-specific configuration settings.
type CrawlerConfig struct {
	// BaseURL is the starting point for the crawler
	BaseURL string
	// MaxDepth defines how many levels deep the crawler should traverse
	MaxDepth int
	// RateLimit defines the delay between requests
	RateLimit time.Duration
	// RandomDelay adds randomization to the delay between requests
	RandomDelay time.Duration
	// IndexName is the Elasticsearch index for storing crawled content
	IndexName string
	// ContentIndexName is the Elasticsearch index for storing parsed content
	ContentIndexName string
	// SourceFile is the path to the sources configuration file
	SourceFile string
	// Parallelism defines how many concurrent crawlers to run
	Parallelism int
}

// Params holds the parameters for creating a config.
type Params struct {
	Environment string
	Debug       bool
	Command     string
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

func (c *NoOpConfig) GetSources() ([]Source, error) {
	return []Source{}, nil
}

func (c *NoOpConfig) GetCommand() string {
	return "test"
}

func (c *NoOpConfig) GetCrawlerConfig() *CrawlerConfig {
	return &CrawlerConfig{
		MaxDepth:    DefaultParallelism,
		RateLimit:   time.Second * DefaultParallelism,
		RandomDelay: time.Second,
		Parallelism: DefaultParallelism,
	}
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *CrawlerConfig {
	return &CrawlerConfig{
		MaxDepth:    crawler.DefaultParallelism,
		RateLimit:   time.Second * crawler.DefaultParallelism,
		Parallelism: crawler.DefaultParallelism,
	}
}
