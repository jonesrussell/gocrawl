// Package config provides configuration management for the GoCrawl application.
// This file specifically handles dependency injection and module initialization
// using the fx framework.
package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	// defaultRetryMaxWait is the default maximum wait time between retries
	defaultRetryMaxWait = 30 * time.Second

	// defaultRetryInitialWait is the default initial wait time between retries
	defaultRetryInitialWait = 1 * time.Second

	// defaultMaxRetries is the default number of retries for failed requests
	defaultMaxRetries = 3
)

// config implements the Interface and holds the actual configuration values
type config struct {
	App           AppConfig           `yaml:"app"`
	Crawler       CrawlerConfig       `yaml:"crawler"`
	Elasticsearch ElasticsearchConfig `yaml:"elasticsearch"`
	Log           LogConfig           `yaml:"log"`
	Sources       []Source            `yaml:"sources"`
	Server        ServerConfig        `yaml:"server"`
}

// Ensure config implements Interface
var _ Interface = (*config)(nil)

// GetCrawlerConfig implements Interface
func (c *config) GetCrawlerConfig() *CrawlerConfig {
	return &c.Crawler
}

// GetElasticsearchConfig implements Interface
func (c *config) GetElasticsearchConfig() *ElasticsearchConfig {
	return &c.Elasticsearch
}

// GetLogConfig implements Interface
func (c *config) GetLogConfig() *LogConfig {
	return &c.Log
}

// GetAppConfig implements Interface
func (c *config) GetAppConfig() *AppConfig {
	return &c.App
}

// GetSources implements Interface
func (c *config) GetSources() []Source {
	return c.Sources
}

// GetServerConfig implements Interface
func (c *config) GetServerConfig() *ServerConfig {
	return &c.Server
}

// InitializeConfig sets up the configuration for the application.
// It handles loading configuration from files and environment variables,
// setting default values, and validating the configuration.
//
// Returns:
//   - Interface: The initialized configuration
//   - error: Any error that occurred during initialization
func InitializeConfig() (Interface, error) {
	// Set config defaults if not already configured
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Set default values for essential configuration
	viper.SetDefault("log.level", "info")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("crawler.source_file", "sources.yml")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError *viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Module provides the config module and its dependencies using fx.
// It sets up the configuration providers that can be used throughout
// the application for dependency injection.
//
// The module provides:
// - Interface instance via the New constructor
// - HTTP transport configuration via NewHTTPTransport
var Module = fx.Options(
	fx.Provide(
		New,
		NewHTTPTransport, // Provides HTTP transport configuration
	),
)

// New creates a new Config instance with values from Viper.
// It handles loading configuration from files, environment variables,
// and setting up default values. It also performs validation of the
// configuration values.
//
// Returns:
//   - Interface: The new configuration instance
//   - error: Any error that occurred during creation
func New() (Interface, error) {
	// Set config defaults if not already configured
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Set default values for essential configuration
	viper.SetDefault("log.level", "info")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("crawler.source_file", "sources.yml")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError *viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse and validate the rate limit configuration
	rateLimit, err := parseRateLimit(viper.GetString("crawler.rate_limit"))
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit: %w", err)
	}

	// Create the configuration instance with values from Viper
	cfg := &config{
		App: AppConfig{
			Environment: viper.GetString("app.environment"),
			Name:        viper.GetString("app.name"),
			Version:     viper.GetString("app.version"),
		},
		Crawler: CrawlerConfig{
			BaseURL:          viper.GetString("crawler.base_url"),
			MaxDepth:         viper.GetInt("crawler.max_depth"),
			RateLimit:        rateLimit,
			RandomDelay:      viper.GetDuration("crawler.random_delay"),
			IndexName:        viper.GetString("elasticsearch.index_name"),
			ContentIndexName: viper.GetString("elasticsearch.content_index_name"),
			SourceFile:       viper.GetString("crawler.source_file"),
			Parallelism:      viper.GetInt("crawler.parallelism"),
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses: viper.GetStringSlice("elasticsearch.addresses"),
			Username:  viper.GetString("elasticsearch.username"),
			Password:  viper.GetString("elasticsearch.password"),
			APIKey:    viper.GetString("elasticsearch.api_key"),
			IndexName: viper.GetString("elasticsearch.index_name"),
			TLS: struct {
				Enabled     bool   `yaml:"enabled"`
				SkipVerify  bool   `yaml:"skip_verify"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
				CA          string `yaml:"ca"`
			}{
				Enabled:     viper.GetBool("elasticsearch.tls.enabled"),
				SkipVerify:  viper.GetBool("elasticsearch.tls.skip_verify"),
				Certificate: viper.GetString("elasticsearch.tls.certificate"),
				Key:         viper.GetString("elasticsearch.tls.key"),
				CA:          viper.GetString("elasticsearch.tls.ca"),
			},
			Retry: struct {
				Enabled     bool          `yaml:"enabled"`
				InitialWait time.Duration `yaml:"initial_wait"`
				MaxWait     time.Duration `yaml:"max_wait"`
				MaxRetries  int           `yaml:"max_retries"`
			}{
				Enabled:     viper.GetBool("elasticsearch.retry.enabled"),
				InitialWait: viper.GetDuration("elasticsearch.retry.initial_wait"),
				MaxWait:     viper.GetDuration("elasticsearch.retry.max_wait"),
				MaxRetries:  viper.GetInt("elasticsearch.retry.max_retries"),
			},
		},
		Log: LogConfig{
			Level: viper.GetString("log.level"),
			Debug: viper.GetBool("log.debug"),
		},
		Server: ServerConfig{
			Address:      viper.GetString("server.address"),
			ReadTimeout:  viper.GetDuration("server.read_timeout"),
			WriteTimeout: viper.GetDuration("server.write_timeout"),
			IdleTimeout:  viper.GetDuration("server.idle_timeout"),
		},
	}

	// Validate the configuration before returning
	if validateErr := ValidateConfig(cfg); validateErr != nil {
		return nil, fmt.Errorf("config validation failed: %w", validateErr)
	}

	return cfg, nil
}

// ProvideConfig provides the configuration with default values.
// This is used internally by InitializeConfig.
func ProvideConfig() (*Config, error) {
	cfg := &Config{
		Elasticsearch: ElasticsearchConfig{
			Retry: struct {
				Enabled     bool          `yaml:"enabled"`
				InitialWait time.Duration `yaml:"initial_wait"`
				MaxWait     time.Duration `yaml:"max_wait"`
				MaxRetries  int           `yaml:"max_retries"`
			}{
				Enabled:     true,
				InitialWait: defaultRetryInitialWait,
				MaxWait:     defaultRetryMaxWait,
				MaxRetries:  defaultMaxRetries,
			},
		},
		Server: ServerConfig{
			Address:      ":8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
	return cfg, nil
}
