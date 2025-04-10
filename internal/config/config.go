// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	logconfig "github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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
	// Sources holds the list of sources to crawl
	Sources []types.Source `yaml:"sources"`
	// Command is the current command being executed
	Command string `yaml:"command"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server: %w", err)
	}
	if err := c.Priority.Validate(); err != nil {
		return fmt.Errorf("priority: %w", err)
	}
	if err := c.Storage.Validate(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	if err := c.Crawler.Validate(); err != nil {
		return fmt.Errorf("crawler: %w", err)
	}
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	if err := c.Elasticsearch.Validate(); err != nil {
		return fmt.Errorf("elasticsearch: %w", err)
	}
	for i := range c.Sources {
		if err := c.Sources[i].Validate(); err != nil {
			return fmt.Errorf("source[%d]: %w", i, err)
		}
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	// Elasticsearch defaults
	v.SetDefault("elasticsearch.bulk_size", DefaultBulkSize)
	v.SetDefault("elasticsearch.flush_interval", DefaultFlushInterval)
	v.SetDefault("elasticsearch.discover_nodes", true)
	v.SetDefault("elasticsearch.retry.enabled", true)
	v.SetDefault("elasticsearch.retry.initial_wait", DefaultRetryInitialWait)
	v.SetDefault("elasticsearch.retry.max_wait", DefaultRetryMaxWait)
	v.SetDefault("elasticsearch.retry.max_retries", DefaultMaxRetries)
	v.SetDefault("elasticsearch.tls.enabled", false)
	v.SetDefault("elasticsearch.tls.insecure_skip_verify", false)
}

func bindEnvVars(v *viper.Viper) error {
	envVars := map[string]string{
		"elasticsearch.api_key":                  "ELASTICSEARCH_API_KEY",
		"elasticsearch.addresses":                "ELASTICSEARCH_HOSTS",
		"elasticsearch.index_name":               "ELASTICSEARCH_INDEX_PREFIX",
		"elasticsearch.username":                 "ELASTIC_USERNAME",
		"elasticsearch.password":                 "ELASTIC_PASSWORD",
		"elasticsearch.cloud.id":                 "ELASTICSEARCH_CLOUD_ID",
		"elasticsearch.cloud.api_key":            "ELASTICSEARCH_CLOUD_API_KEY",
		"elasticsearch.tls.enabled":              "ELASTIC_TLS_ENABLED",
		"elasticsearch.tls.cert_file":            "ELASTIC_TLS_CERT_FILE",
		"elasticsearch.tls.key_file":             "ELASTIC_TLS_KEY_FILE",
		"elasticsearch.tls.ca_file":              "ELASTIC_TLS_CA_FILE",
		"elasticsearch.tls.insecure_skip_verify": "ELASTICSEARCH_SKIP_TLS",
		"elasticsearch.retry.enabled":            "ELASTICSEARCH_RETRY_ENABLED",
		"elasticsearch.retry.initial_wait":       "ELASTICSEARCH_RETRY_INITIAL_WAIT",
		"elasticsearch.retry.max_wait":           "ELASTICSEARCH_RETRY_MAX_WAIT",
		"elasticsearch.retry.max_retries":        "ELASTICSEARCH_MAX_RETRIES",
		"elasticsearch.bulk_size":                "ELASTICSEARCH_BULK_SIZE",
		"elasticsearch.flush_interval":           "ELASTICSEARCH_FLUSH_INTERVAL",
		"elasticsearch.discover_nodes":           "ELASTICSEARCH_DISCOVER_NODES",
	}

	for configKey, envVar := range envVars {
		if err := v.BindEnv(configKey, envVar); err != nil {
			return fmt.Errorf("failed to bind environment variable %s: %w", envVar, err)
		}
	}

	return nil
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	var loadErr error

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(path)
	v.AddConfigPath(".")

	// Set default values
	setDefaults(v)

	// Load environment variables
	if loadErr = bindEnvVars(v); loadErr != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", loadErr)
	}

	// Load .env files
	if loadErr = godotenv.Load(".env", ".env.development"); loadErr != nil && !os.IsNotExist(loadErr) {
		return nil, fmt.Errorf("failed to load .env files: %w", loadErr)
	}

	// Read config file (if it exists)
	if loadErr = v.ReadInConfig(); loadErr != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(loadErr, &configFileNotFound) {
			return nil, fmt.Errorf("failed to read config file: %w", loadErr)
		}
	}

	// Unmarshal config
	if loadErr = v.Unmarshal(&cfg); loadErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", loadErr)
	}

	// Validate config
	if loadErr = cfg.Validate(); loadErr != nil {
		return nil, fmt.Errorf("invalid configuration: %w", loadErr)
	}

	return &cfg, nil
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

// GetSources returns the list of sources.
func (c *Config) GetSources() []types.Source {
	return c.Sources
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
