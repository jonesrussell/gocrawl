// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/commands"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/logging"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

const (
	// DefaultMaxLogSize is the default maximum size of a log file in MB
	DefaultMaxLogSize = 100
	// DefaultMaxLogBackups is the default number of log file backups to keep
	DefaultMaxLogBackups = 3
	// DefaultMaxLogAge is the default maximum age of a log file in days
	DefaultMaxLogAge = 30
)

// Config represents the application configuration.
type Config struct {
	// Environment is the application environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Logger holds logging-specific configuration
	Logger *logging.Config `yaml:"logger"`
	// Server holds server-specific configuration
	Server *server.Config `yaml:"server"`
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
	// logger is the application logger
	logger logger.Interface
}

// NewConfig creates a new config instance.
func NewConfig(logger logger.Interface) *Config {
	return &Config{
		logger: logger,
	}
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

// LoadConfig loads the configuration from Viper into a Config struct.
func LoadConfig() (*Config, error) {
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
	if unmarshalErr := viper.Unmarshal(&cfg); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", unmarshalErr)
	}

	// Set the logger
	cfg.logger = tempLogger

	// Validate the configuration
	if validateErr := cfg.Validate(); validateErr != nil {
		return nil, fmt.Errorf("invalid config: %w", validateErr)
	}

	return &cfg, nil
}

// GetAppConfig returns the application configuration.
func (c *Config) GetAppConfig() *app.Config {
	return c.App
}

// GetLogConfig returns the logging configuration.
func (c *Config) GetLogConfig() *logging.Config {
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
