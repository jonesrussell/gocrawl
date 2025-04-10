// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/commands"
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

// Validate validates the configuration based on the current command.
func (c *Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}

	// Always validate logger as it's used by all commands
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}

	// Validate command-specific components
	switch c.Command {
	case commands.IndicesList, commands.IndicesDelete:
		// Only need basic Elasticsearch connection
		if err := c.Elasticsearch.ValidateConnection(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}

	case commands.IndicesCreate:
		// Need Elasticsearch with index name
		if err := c.Elasticsearch.Validate(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}

	case commands.Crawl:
		// Need full Elasticsearch, crawler, and sources
		if err := c.Elasticsearch.Validate(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}
		if c.Crawler == nil {
			return errors.New("crawler configuration is required")
		}
		if err := c.Crawler.Validate(); err != nil {
			return fmt.Errorf("crawler: %w", err)
		}
		if len(c.Sources) == 0 {
			return errors.New("at least one source is required")
		}
		for i := range c.Sources {
			if err := c.Sources[i].Validate(); err != nil {
				return fmt.Errorf("source[%d]: %w", i, err)
			}
		}

	case commands.HTTPD:
		// Need server and storage config
		if err := c.Server.Validate(); err != nil {
			return fmt.Errorf("server: %w", err)
		}
		if err := c.Storage.Validate(); err != nil {
			return fmt.Errorf("storage: %w", err)
		}

	case commands.Search:
		// Need Elasticsearch and storage
		if err := c.Elasticsearch.Validate(); err != nil {
			return fmt.Errorf("elasticsearch: %w", err)
		}
		if err := c.Storage.Validate(); err != nil {
			return fmt.Errorf("storage: %w", err)
		}

	case commands.Sources:
		// Need storage for source management
		if err := c.Storage.Validate(); err != nil {
			return fmt.Errorf("storage: %w", err)
		}
	}

	return nil
}

// LoadConfig loads the configuration from the given path.
func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(".")

	// Set defaults
	setDefaults(v)

	// Load environment
	if err := loadEnvironment(); err != nil {
		return nil, err
	}

	// Bind environment variables
	if err := bindEnvVars(v); err != nil {
		return nil, err
	}

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Ensure TLS config is initialized
	ensureTLSConfig(&cfg, v)

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")
	v.SetDefault("elasticsearch.addresses", []string{"https://localhost:9200"})
	v.SetDefault("elasticsearch.tls.insecure_skip_verify", true)
}

// loadEnvironment loads environment variables from .env file
func loadEnvironment() error {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Failed to load .env file: %v\n", err)
	}
	return nil
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

// ensureTLSConfig ensures TLS configuration is properly initialized
func ensureTLSConfig(cfg *Config, v *viper.Viper) {
	if cfg.Elasticsearch.TLS == nil {
		cfg.Elasticsearch.TLS = &elasticsearch.TLSConfig{
			InsecureSkipVerify: true,
		}
	}

	// Force TLS settings from environment
	if v.IsSet("elasticsearch.tls.insecure_skip_verify") {
		cfg.Elasticsearch.TLS.InsecureSkipVerify = v.GetBool("elasticsearch.tls.insecure_skip_verify")
	}
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
