// Package config provides configuration management for the GoCrawl application.
// It handles loading, validation, and access to configuration values from both
// YAML files and environment variables using Viper.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	logconfig "github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/config/storage"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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
	// Sources holds the list of sources to crawl
	Sources []types.Source `yaml:"sources"`
	// Command is the current command being executed
	Command string `yaml:"command"`
}

// newConfig returns a new configuration with default values.
func newConfig() *Config {
	esConfig := elasticsearch.NewConfig()
	esConfig.IndexName = "gocrawl"
	esConfig.TLS.Enabled = false // Disable TLS by default for development

	return &Config{
		Environment:   "development",
		Logger:        logconfig.NewConfig(),
		Server:        server.NewConfig(),
		Priority:      priority.NewConfig(),
		Storage:       storage.NewConfig(),
		Crawler:       crawler.New(crawler.WithBaseURL("https://example.com")),
		App:           app.New(),
		Elasticsearch: esConfig,
	}
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

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	// Create a logger for configuration
	log, err := logger.New(&logger.Config{
		Level:       logger.InfoLevel,
		Development: true,
		Encoding:    "console",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Load .env files
	if err := godotenv.Load(".env", ".env.development"); err != nil {
		log.Warn("Error loading .env files", "error", err)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Replace dots with underscores in env variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Enable reading env variables
	v.AutomaticEnv()

	// Allow empty environment variables for optional fields
	v.AllowEmptyEnv(true)

	// Bind environment variables to configuration fields
	v.BindEnv("elasticsearch.api_key", "ELASTICSEARCH_API_KEY")
	v.BindEnv("elasticsearch.addresses", "ELASTICSEARCH_HOSTS")
	v.BindEnv("elasticsearch.index_name", "ELASTICSEARCH_INDEX_PREFIX")
	v.BindEnv("elasticsearch.username", "ELASTIC_USERNAME")
	v.BindEnv("elasticsearch.password", "ELASTIC_PASSWORD")
	v.BindEnv("elasticsearch.cloud.id", "ELASTICSEARCH_CLOUD_ID")
	v.BindEnv("elasticsearch.cloud.api_key", "ELASTICSEARCH_CLOUD_API_KEY")
	v.BindEnv("elasticsearch.tls.enabled", "ELASTIC_TLS_ENABLED")
	v.BindEnv("elasticsearch.tls.cert_file", "ELASTIC_TLS_CERT_FILE")
	v.BindEnv("elasticsearch.tls.key_file", "ELASTIC_TLS_KEY_FILE")
	v.BindEnv("elasticsearch.tls.ca_file", "ELASTIC_TLS_CA_FILE")
	v.BindEnv("elasticsearch.tls.insecure_skip_verify", "ELASTICSEARCH_SKIP_TLS")
	v.BindEnv("elasticsearch.retry.enabled", "ELASTICSEARCH_RETRY_ENABLED")
	v.BindEnv("elasticsearch.retry.initial_wait", "ELASTICSEARCH_RETRY_INITIAL_WAIT")
	v.BindEnv("elasticsearch.retry.max_wait", "ELASTICSEARCH_RETRY_MAX_WAIT")
	v.BindEnv("elasticsearch.retry.max_retries", "ELASTICSEARCH_MAX_RETRIES")
	v.BindEnv("elasticsearch.bulk_size", "ELASTICSEARCH_BULK_SIZE")
	v.BindEnv("elasticsearch.flush_interval", "ELASTICSEARCH_FLUSH_INTERVAL")
	v.BindEnv("elasticsearch.discover_nodes", "ELASTICSEARCH_DISCOVER_NODES")

	// Read config file (if it exists)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, using defaults and environment variables
		log.Info("No config file found, using defaults and environment variables")
	}

	// Get environment to determine logging level
	env := v.GetString("app.environment")
	if env == "" {
		env = "development"
	}

	// Only log configuration in development mode
	if env == "development" {
		log.Info("Loading configuration", "environment", env)

		// Log environment variables that are set
		configKeys := []string{
			"app.name",
			"app.env",
			"app.debug",
			"server.port",
			"server.read_timeout",
			"server.write_timeout",
			"server.idle_timeout",
			"elasticsearch.addresses",
			"elasticsearch.api_key",
			"elasticsearch.index_name",
			"elasticsearch.max_retries",
			"elasticsearch.retry_initial_wait",
			"elasticsearch.retry_max_wait",
			"crawler.max_depth",
			"crawler.parallelism",
			"crawler.max_age",
			"crawler.rate_limit",
			"log.level",
			"log.format",
			"elastic.username",
			"elastic.password",
			"elastic.skip_tls",
			"elastic.ca_fingerprint",
			"gocrawl.port",
			"gocrawl.api_key",
		}

		log.Info("Environment Variables:")
		for _, key := range configKeys {
			if v.IsSet(key) {
				// Mask sensitive values in the log message
				if strings.Contains(key, "api_key") || strings.Contains(key, "password") {
					log.Info("  "+strings.ReplaceAll(key, ".", "_"), "value", "[MASKED]")
				} else {
					log.Info("  "+strings.ReplaceAll(key, ".", "_"), "value", v.Get(key))
				}
			}
		}
	}

	cfg := newConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Explicitly set API key from environment variable
	if apiKey := v.GetString("ELASTICSEARCH_API_KEY"); apiKey != "" {
		cfg.Elasticsearch.APIKey = apiKey
	}

	// Explicitly set TLS skip verify from environment variable
	if skipTLS := v.GetBool("ELASTICSEARCH_SKIP_TLS"); skipTLS {
		cfg.Elasticsearch.TLS.InsecureSkipVerify = true
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
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
