// Package config provides configuration management for the GoCrawl application.
// This file specifically handles dependency injection and module initialization
// using the fx framework.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// InitializeConfig sets up the configuration for the application.
// It handles loading configuration from files and environment variables,
// setting default values, and validating the configuration.
//
// Parameters:
//   - cfgFile: Path to the configuration file (optional)
//
// Returns:
//   - *Config: The initialized configuration
//   - error: Any error that occurred during initialization
func InitializeConfig(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	// Set default values for essential configuration
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("APP_ENV", "development")

	// Bind environment variables and check for errors
	if err := viper.BindEnv("LOG_LEVEL"); err != nil {
		return nil, fmt.Errorf("failed to bind LOG_LEVEL environment variable: %w", err)
	}
	if err := viper.BindEnv("APP_ENV"); err != nil {
		return nil, fmt.Errorf("failed to bind APP_ENV environment variable: %w", err)
	}

	return New()
}

// New creates a new Config instance with values from Viper.
// It handles loading configuration from files, environment variables,
// and setting up default values. It also performs validation of the
// configuration values.
//
// Returns:
//   - *Config: The new configuration instance
//   - error: Any error that occurred during creation
func New() (*Config, error) {
	// Set config defaults if not already configured
	if viper.ConfigFileUsed() == "" {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configErr *viper.ConfigFileNotFoundError
		if errors.As(err, &configErr) {
			// Log to stderr instead of using fmt.Println for better error handling
			fmt.Fprintf(os.Stderr, "Config file not found; using environment variables\n")
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Parse and validate the rate limit configuration
	rateLimit, err := parseRateLimit(viper.GetString(CrawlerRateLimitKey))
	if err != nil {
		return nil, fmt.Errorf("error parsing rate limit: %w", err)
	}

	// Create the configuration instance with values from Viper
	cfg := &Config{
		App: AppConfig{
			Environment: viper.GetString(AppEnvKey),
			Name:        viper.GetString("APP_NAME"),
			Version:     viper.GetString("APP_VERSION"),
		},
		Crawler: CrawlerConfig{
			BaseURL:          viper.GetString(CrawlerBaseURLKey),
			MaxDepth:         viper.GetInt(CrawlerMaxDepthKey),
			RateLimit:        rateLimit,
			RandomDelay:      viper.GetDuration("CRAWLER_RANDOM_DELAY"),
			IndexName:        viper.GetString(ElasticIndexNameKey),
			ContentIndexName: viper.GetString("ELASTIC_CONTENT_INDEX_NAME"),
			SourceFile:       viper.GetString(CrawlerSourceFileKey),
			Parallelism:      viper.GetInt("CRAWLER_PARALLELISM"),
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses: []string{viper.GetString(ElasticURLKey)},
			Username:  viper.GetString(ElasticUsernameKey),
			Password:  viper.GetString(ElasticPasswordKey),
			APIKey:    viper.GetString(ElasticAPIKeyKey),
			IndexName: viper.GetString(ElasticIndexNameKey),
			TLS: struct {
				Enabled     bool   `yaml:"enabled"`
				SkipVerify  bool   `yaml:"skip_verify"`
				Certificate string `yaml:"certificate"`
				Key         string `yaml:"key"`
				CA          string `yaml:"ca"`
			}{
				Enabled:    true,
				SkipVerify: viper.GetBool(ElasticSkipTLSKey),
			},
			Retry: struct {
				Enabled     bool          `yaml:"enabled"`
				InitialWait time.Duration `yaml:"initial_wait"`
				MaxWait     time.Duration `yaml:"max_wait"`
				MaxRetries  int           `yaml:"max_retries"`
			}{
				Enabled:     true,
				InitialWait: time.Second,
				MaxWait:     time.Second * 30,
				MaxRetries:  3,
			},
		},
		Log: LogConfig{
			Level: viper.GetString(LogLevelKey),
			Debug: viper.GetBool(AppDebugKey),
		},
	}

	// Validate the configuration before returning
	if validateErr := ValidateConfig(cfg); validateErr != nil {
		return nil, fmt.Errorf("config validation failed: %w", validateErr)
	}

	return cfg, nil
}

// Module provides the config module and its dependencies using fx.
// It sets up the configuration providers that can be used throughout
// the application for dependency injection.
//
// The module provides:
// - Config instance via the New constructor
// - HTTP transport configuration via NewHTTPTransport
var Module = fx.Options(
	fx.Provide(
		New,              // Provides the configuration instance
		NewHTTPTransport, // Provides HTTP transport configuration
	),
)
