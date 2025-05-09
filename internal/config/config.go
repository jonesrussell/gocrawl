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

	// Load .env file if it exists
	v.SetConfigFile(".env")
	v.MergeInConfig()

	// Set defaults
	setDefaults(v)

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// If config file is not found, we'll use the defaults set by setDefaults
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
	if unmarshalErr := v.Unmarshal(&cfg); unmarshalErr != nil {
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

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app", map[string]any{
		"name":        "gocrawl",
		"version":     "1.0.0",
		"environment": "development",
		"debug":       true,
	})

	// Logger defaults
	v.SetDefault("logger", map[string]any{
		"level":       "debug",
		"encoding":    "console",
		"output":      "stdout",
		"debug":       true,
		"caller":      false,
		"stacktrace":  false,
		"max_size":    DefaultMaxLogSize,
		"max_backups": DefaultMaxLogBackups,
		"max_age":     DefaultMaxLogAge,
		"compress":    true,
	})

	// Crawler defaults
	v.SetDefault("crawler", map[string]any{
		"max_depth":   DefaultMaxDepth,
		"max_retries": DefaultMaxRetries,
		"rate_limit":  "1s",
		"timeout":     "30s",
		"user_agent":  "GoCrawl/1.0",
		"source_file": "sources.yml",
	})

	// Storage defaults
	v.SetDefault("storage", map[string]any{
		"type":           "elasticsearch",
		"batch_size":     DefaultStorageBatchSize,
		"flush_interval": "5s",
	})

	// Elasticsearch defaults
	v.SetDefault("elasticsearch", map[string]any{
		"addresses":  []string{"https://localhost:9200"},
		"index_name": "gocrawl",
		"retry": map[string]any{
			"enabled":      true,
			"initial_wait": "1s",
			"max_wait":     "5s",
			"max_retries":  DefaultElasticsearchRetries,
		},
		"bulk_size":      DefaultBulkSize,
		"flush_interval": "1s",
		"tls": map[string]any{
			"insecure_skip_verify": true,
		},
	})
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
