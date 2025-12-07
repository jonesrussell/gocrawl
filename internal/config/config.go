// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/commands"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/logging"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

// Interface defines the interface for configuration management.
type Interface interface {
	// GetAppConfig returns the application configuration.
	GetAppConfig() *app.Config
	// GetLogConfig returns the logging configuration.
	GetLogConfig() *logging.Config
	// GetServerConfig returns the server configuration.
	GetServerConfig() *server.Config
	// GetCrawlerConfig returns the crawler configuration.
	GetCrawlerConfig() *crawler.Config
	// GetElasticsearchConfig returns the Elasticsearch configuration.
	GetElasticsearchConfig() *elasticsearch.Config
	// GetCommand returns the current command.
	GetCommand() string
	// GetConfigFile returns the path to the configuration file.
	GetConfigFile() string
	// Validate validates the configuration based on the current command.
	Validate() error
}

// Default configuration values
const (
	// DefaultMaxLogSize is the default maximum size of a log file in MB
	DefaultMaxLogSize = 100
	// DefaultMaxLogBackups is the default number of log file backups to keep
	DefaultMaxLogBackups = 3
	// DefaultMaxLogAge is the default maximum age of a log file in days
	DefaultMaxLogAge = 30

	// Server defaults
	DefaultServerAddress      = ":8080"
	DefaultServerReadTimeout  = 30 * time.Second
	DefaultServerWriteTimeout = 30 * time.Second
	DefaultServerIdleTimeout  = 60 * time.Second
)

// Ensure Config implements Interface
var _ Interface = (*Config)(nil)

// Config represents the application configuration.
type Config struct {
	// Environment is the application environment (development, staging, production)
	Environment string `yaml:"environment"`
	// Logger holds logging-specific configuration
	Logger *logging.Config `yaml:"logger"`
	// Server holds server-specific configuration
	Server *server.Config `yaml:"server"`
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
func NewConfig(log logger.Interface) *Config {
	return &Config{
		logger: log,
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
	if err := c.Elasticsearch.Validate(); err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
	}
	return nil
}

// validateSearchConfig validates the configuration for the search command
func (c *Config) validateSearchConfig() error {
	if err := c.Elasticsearch.Validate(); err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
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
		if err := c.Elasticsearch.Validate(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}
	}

	return nil
}

// LoadConfig loads the configuration from Viper
func LoadConfig() (*Config, error) {
	// Create a temporary logger for config loading
	// Use the logger level from Viper if available, otherwise default to info
	logLevel := logger.InfoLevel
	if viper.IsSet("logger.level") {
		logLevel = logger.Level(viper.GetString("logger.level"))
	} else if viper.GetBool("app.debug") {
		logLevel = logger.DebugLevel
	}

	tempLogger, err := logger.New(&logger.Config{
		Level:       logLevel,
		Development: viper.GetBool("logger.development") || viper.GetBool("app.debug"),
		Encoding:    "console",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary logger: %w", err)
	}

	// Create config with defaults
	cfg := &Config{
		Environment: viper.GetString("environment"),
		Logger: &logging.Config{
			Level:    viper.GetString("logger.level"),
			Encoding: viper.GetString("logger.encoding"),
		},
		Server:        server.NewConfig(),
		Elasticsearch: elasticsearch.LoadFromViper(viper.GetViper()),
		Crawler:       crawler.LoadFromViper(viper.GetViper()),
		App: &app.Config{
			Name:        viper.GetString("app.name"),
			Version:     viper.GetString("app.version"),
			Environment: viper.GetString("app.environment"),
			Debug:       viper.GetBool("app.debug"),
		},
	}

	// Set default app values if not set
	if cfg.App.Name == "" {
		cfg.App.Name = "gocrawl"
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "1.0.0"
	}
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}

	// Set server config from Viper with defaults
	cfg.Server.Address = viper.GetString("server.address")
	if cfg.Server.Address == "" {
		cfg.Server.Address = DefaultServerAddress
	}

	cfg.Server.ReadTimeout = viper.GetDuration("server.readTimeout")
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = DefaultServerReadTimeout
	}

	cfg.Server.WriteTimeout = viper.GetDuration("server.writeTimeout")
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = DefaultServerWriteTimeout
	}

	cfg.Server.IdleTimeout = viper.GetDuration("server.idleTimeout")
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = DefaultServerIdleTimeout
	}

	cfg.Server.SecurityEnabled = viper.GetBool("server.security.enabled")
	cfg.Server.APIKey = viper.GetString("server.security.apiKey")

	// Set the logger
	cfg.logger = tempLogger

	// Validate the configuration
	if validateErr := cfg.Validate(); validateErr != nil {
		return nil, fmt.Errorf("invalid config: %w", validateErr)
	}

	return cfg, nil
}

// LoadElasticsearchConfig loads Elasticsearch configuration from Viper
func LoadElasticsearchConfig() *elasticsearch.Config {
	return elasticsearch.LoadFromViper(viper.GetViper())
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

// GetConfigFile returns the path to the configuration file.
func (c *Config) GetConfigFile() string {
	return viper.ConfigFileUsed()
}
