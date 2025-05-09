// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/commands"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	logconfig "github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	// Environment is the application environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Logger holds logging-specific configuration
	Logger *logconfig.Config `yaml:"logger"`
	// Server holds server-specific configuration
	Server *server.Config `yaml:"server"`
	// Priority holds priority-specific configuration
	Priority *priority.Config `yaml:"priority"`
	// Storage holds storage-specific configuration
	Storage *storage.Config `yaml:"storage"`
	// Crawler holds crawler-specific configuration
	Crawler *crawler.Config `yaml:"crawler"`
	// App holds application-specific configuration
	App *app.Config `yaml:"app"`
	// Elasticsearch holds Elasticsearch configuration
	Elasticsearch *elasticsearch.Config `yaml:"elasticsearch"`
	// Command is the current command being executed
	Command string `yaml:"command"`
	viper   *viper.Viper
	logger  logger.Interface
}

// NewConfig creates a new config instance.
func NewConfig(logger logger.Interface) *Config {
	return &Config{
		viper:  viper.New(),
		logger: logger,
	}
}

// validateBasicConfig validates the basic configuration that's common to all commands
func (c *Config) validateBasicConfig() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}

	// Always validate logger as it's used by all commands
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}

	return nil
}

// validateCrawlConfig validates the configuration for the crawl command
func (c *Config) validateCrawlConfig() error {
	if err := c.Elasticsearch.Validate(); err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
	}
	if c.Crawler == nil {
		return errors.New("crawler configuration is required")
	}
	if err := c.Crawler.Validate(); err != nil {
		return fmt.Errorf("crawler: %w", err)
	}
	return nil
}

// validateHTTPDConfig validates the configuration for the httpd command
func (c *Config) validateHTTPDConfig() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	return nil
}

// validateSearchConfig validates the configuration for the search command
func (c *Config) validateSearchConfig() error {
	if err := c.Elasticsearch.Validate(); err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	return nil
}

// Validate validates the configuration based on the current command.
func (c *Config) Validate() error {
	if err := c.validateBasicConfig(); err != nil {
		return err
	}

	switch c.Command {
	case commands.IndicesList, commands.IndicesDelete, commands.IndicesCreate:
		if err := c.Elasticsearch.Validate(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}

	case commands.Crawl:
		if err := c.validateCrawlConfig(); err != nil {
			return err
		}

	case commands.HTTPD:
		if err := c.validateHTTPDConfig(); err != nil {
			return err
		}

	case commands.Search:
		if err := c.validateSearchConfig(); err != nil {
			return err
		}

	case commands.Sources:
		if err := c.Storage.Validate(); err != nil {
			return fmt.Errorf("storage: %w", err)
		}
	}

	return nil
}

// LoadConfig loads the configuration from the specified file or default locations.
func LoadConfig() (*Config, error) {
	// Create a new Viper instance
	v := viper.New()

	// Set up Viper
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	// Add config paths in order of priority
	v.AddConfigPath(".")              // Current directory
	v.AddConfigPath("$HOME/.gocrawl") // User config directory
	v.AddConfigPath("/etc/gocrawl")   // System config directory

	// Set up environment variable binding
	v.AutomaticEnv()
	v.SetEnvPrefix("GOCRAWL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// If config file is not found, create a default config
		v.Set("app.name", "gocrawl")
		v.Set("app.environment", "development")
		v.Set("app.debug", true)
		v.Set("logger.level", "debug")
		v.Set("logger.encoding", "console")
		v.Set("logger.format", "text")
		v.Set("logger.output", "stdout")
		v.Set("crawler.source_file", "sources.yml")
		v.Set("crawler.max_depth", DefaultMaxDepth)
		v.Set("crawler.max_retries", DefaultMaxRetries)
		v.Set("storage.batch_size", DefaultStorageBatchSize)
		v.Set("elasticsearch.max_retries", DefaultElasticsearchRetries)
	}

	// Create a temporary logger for config loading
	tempLogger, err := logger.New(&logger.Config{
		Level:       logger.InfoLevel,
		Development: true,
		Encoding:    "console",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary logger: %w", err)
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set the logger
	cfg.logger = tempLogger

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "gocrawl")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", false)

	// Logger defaults
	v.SetDefault("logger.level", "debug")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.enable_color", true)

	// Crawler defaults
	v.SetDefault("crawler.max_depth", DefaultMaxDepth)
	v.SetDefault("crawler.max_retries", DefaultMaxRetries)
	v.SetDefault("crawler.rate_limit", "1s")
	v.SetDefault("crawler.timeout", "30s")
	v.SetDefault("crawler.user_agent", "GoCrawl/1.0")

	// Storage defaults
	v.SetDefault("storage.type", "elasticsearch")
	v.SetDefault("storage.batch_size", DefaultStorageBatchSize)
	v.SetDefault("storage.flush_interval", "5s")

	// Elasticsearch defaults
	v.SetDefault("elasticsearch.addresses", []string{"https://localhost:9200"})
	v.SetDefault("elasticsearch.index_name", "gocrawl")
	v.SetDefault("elasticsearch.retry.enabled", true)
	v.SetDefault("elasticsearch.retry.initial_wait", "1s")
	v.SetDefault("elasticsearch.retry.max_wait", "5s")
	v.SetDefault("elasticsearch.retry.max_retries", DefaultElasticsearchRetries)
	v.SetDefault("elasticsearch.bulk_size", DefaultBulkSize)
	v.SetDefault("elasticsearch.flush_interval", "1s")
	v.SetDefault("elasticsearch.tls.insecure_skip_verify", true)
}

// GetAppConfig returns the application configuration.
func (c *Config) GetAppConfig() *app.Config {
	return c.App
}

// GetLogConfig returns the logging configuration.
func (c *Config) GetLogConfig() *logconfig.Config {
	return c.Logger
}

// GetServerConfig returns the server configuration.
func (c *Config) GetServerConfig() *server.Config {
	return c.Server
}

// GetCrawlerConfig returns the crawler configuration.
func (c *Config) GetCrawlerConfig() *crawler.Config {
	return c.Crawler
}

// GetPriorityConfig returns the priority configuration.
func (c *Config) GetPriorityConfig() *priority.Config {
	return c.Priority
}

// GetElasticsearchConfig returns the Elasticsearch configuration.
func (c *Config) GetElasticsearchConfig() *elasticsearch.Config {
	return c.Elasticsearch
}

// GetCommand returns the current command.
func (c *Config) GetCommand() string {
	return c.Command
}

// GetStorageConfig returns the storage configuration.
func (c *Config) GetStorageConfig() *storage.Config {
	return c.Storage
}

// GetConfigFile returns the path to the configuration file.
func (c *Config) GetConfigFile() string {
	return viper.ConfigFileUsed()
}
