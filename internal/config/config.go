// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/joho/godotenv"
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

// LoadConfig loads the configuration from the given path.
func LoadConfig() (*Config, error) {
	// Use the global Viper instance
	v := viper.GetViper()

	// Set up Viper
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.gocrawl")
	v.AddConfigPath("/etc/gocrawl")

	// Also try with .yaml extension
	v.SetConfigName("config.yaml")
	v.AddConfigPath(".")

	// Set defaults
	setDefaults(v)

	// Load environment
	loadEnvironment()

	// Bind environment variables
	if err := bindEnvVars(v); err != nil {
		return nil, err
	}

	// Read the config file if it exists
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// If config file is not found, create a default config
		v.Set("environment", "development")
		v.Set("logger.level", "debug")
		v.Set("logger.encoding", "console")
		v.Set("logger.format", "text")
		v.Set("logger.output", "stdout")
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
	v.SetConfigType("yaml")
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
	v.SetDefault("environment", "development")
	v.SetDefault("elasticsearch.addresses", []string{"https://localhost:9200"})
	v.SetDefault("elasticsearch.tls.insecure_skip_verify", true)
}

// loadEnvironment loads environment variables from .env file
func loadEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}
}

// bindEnvVars binds environment variables to configuration keys
func bindEnvVars(v *viper.Viper) error {
	envVars := map[string]string{
		"elasticsearch.tls.insecure_skip_verify": "ELASTICSEARCH_TLS_INSECURE_SKIP_VERIFY",
		"elasticsearch.tls.cert_file":            "ELASTICSEARCH_TLS_CERT_FILE",
		"elasticsearch.tls.key_file":             "ELASTICSEARCH_TLS_KEY_FILE",
		"elasticsearch.tls.ca_file":              "ELASTICSEARCH_TLS_CA_FILE",
		"elasticsearch.addresses":                "ELASTICSEARCH_HOSTS",
		"elasticsearch.api_key":                  "ELASTICSEARCH_API_KEY",
		"elasticsearch.username":                 "ELASTICSEARCH_USERNAME",
		"elasticsearch.password":                 "ELASTICSEARCH_PASSWORD",
	}

	for configKey, envVar := range envVars {
		if err := v.BindEnv(configKey, envVar); err != nil {
			return fmt.Errorf("failed to bind environment variable %s: %w", envVar, err)
		}
	}

	return nil
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
