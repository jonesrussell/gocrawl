// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Config holds the complete application configuration.
type Config struct {
	// App holds application-level configuration
	App *AppConfig `yaml:"app"`
	// Log holds logging-related configuration
	Log *LogConfig `yaml:"log"`
	// Crawler holds crawler-specific configuration
	Crawler *CrawlerConfig `yaml:"crawler"`
	// Elasticsearch holds Elasticsearch connection configuration
	Elasticsearch *ElasticsearchConfig `yaml:"elasticsearch"`
}

// ConfigValidationError represents a configuration validation error.
type ConfigValidationError struct {
	Field  string
	Value  interface{}
	Reason string
}

// Error returns a string representation of the validation error.
func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("invalid configuration: %s=%v (%s)", e.Field, e.Value, e.Reason)
}

const (
	envDevelopment = "development"
	envStaging     = "staging"
	envProduction  = "production"
	envTest        = "test"
)

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("configuration is required")
	}

	logger.Debug("Starting config validation",
		zap.String("environment", cfg.App.Environment))

	// Validate environment first
	if cfg.App.Environment == "" {
		logger.Error("Environment validation failed: environment cannot be empty")
		return &ConfigValidationError{
			Field:  "app.environment",
			Value:  cfg.App.Environment,
			Reason: "environment cannot be empty",
		}
	}

	// Validate environment value
	validEnvs := []string{envDevelopment, envStaging, envProduction, envTest}
	isValidEnv := false
	for _, env := range validEnvs {
		if cfg.App.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		logger.Error("Environment validation failed",
			zap.String("environment", cfg.App.Environment))
		return &ConfigValidationError{
			Field:  "app.environment",
			Value:  cfg.App.Environment,
			Reason: "invalid environment",
		}
	}
	logger.Debug("Environment validation passed")

	// Validate log config
	if err := validateLogConfig(cfg.Log); err != nil {
		logger.Error("Log config validation failed", zap.Error(err))
		return err
	}
	logger.Debug("Log config validation passed")

	// Validate crawler config
	if err := validateCrawlerConfig(cfg.Crawler); err != nil {
		logger.Error("Crawler config validation failed", zap.Error(err))
		return err
	}
	logger.Debug("Crawler config validation passed")

	// Validate Elasticsearch config
	if err := validateElasticsearchConfig(cfg.Elasticsearch); err != nil {
		logger.Error("Elasticsearch config validation failed", zap.Error(err))
		return err
	}
	logger.Debug("Elasticsearch config validation passed")

	// Validate app config last
	if err := validateAppConfig(cfg.App); err != nil {
		logger.Error("App config validation failed", zap.Error(err))
		return &ConfigValidationError{
			Field:  "app",
			Value:  cfg.App,
			Reason: err.Error(),
		}
	}
	logger.Debug("App config validation passed")

	// Validate sources last
	if err := validateSources(cfg.Crawler.Sources); err != nil {
		logger.Error("Sources validation failed", zap.Error(err))
		return err
	}
	logger.Debug("Sources validation passed")

	return nil
}

// validateAppConfig validates the application configuration
func validateAppConfig(cfg *AppConfig) error {
	if cfg.Name == "" {
		return &ConfigValidationError{
			Field:  "app.name",
			Value:  cfg.Name,
			Reason: "name cannot be empty",
		}
	}
	if cfg.Version == "" {
		return &ConfigValidationError{
			Field:  "app.version",
			Value:  cfg.Version,
			Reason: "version cannot be empty",
		}
	}
	return nil
}

// validateLogConfig validates the log configuration
func validateLogConfig(cfg *LogConfig) error {
	if cfg == nil {
		return &ConfigValidationError{
			Field:  "log",
			Value:  nil,
			Reason: "log configuration is required",
		}
	}

	if cfg.Level == "" {
		cfg.Level = "info" // Default to info if not set
	}

	if !ValidLogLevels[strings.ToLower(cfg.Level)] {
		return &ConfigValidationError{
			Field:  "log.level",
			Value:  cfg.Level,
			Reason: "invalid log level",
		}
	}
	return nil
}

// validateCrawlerConfig validates the crawler configuration
func validateCrawlerConfig(cfg *CrawlerConfig) error {
	if cfg == nil {
		return &ConfigValidationError{
			Field:  "crawler",
			Value:  nil,
			Reason: "crawler configuration is required",
		}
	}

	if cfg.BaseURL == "" {
		return &ConfigValidationError{
			Field:  "crawler.base_url",
			Value:  cfg.BaseURL,
			Reason: "crawler base URL cannot be empty",
		}
	}

	if cfg.MaxDepth < 1 {
		return &ConfigValidationError{
			Field:  "crawler.max_depth",
			Value:  cfg.MaxDepth,
			Reason: "crawler max depth must be greater than 0",
		}
	}

	if cfg.RateLimit < time.Second {
		return &ConfigValidationError{
			Field:  "crawler.rate_limit",
			Value:  cfg.RateLimit,
			Reason: "crawler rate limit must be at least 1 second",
		}
	}

	if cfg.Parallelism < 1 {
		return &ConfigValidationError{
			Field:  "crawler.parallelism",
			Value:  cfg.Parallelism,
			Reason: "crawler parallelism must be greater than 0",
		}
	}

	if cfg.SourceFile == "" {
		return &ConfigValidationError{
			Field:  "crawler.source_file",
			Value:  cfg.SourceFile,
			Reason: "crawler source file cannot be empty",
		}
	}

	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg *ElasticsearchConfig) error {
	if cfg == nil {
		return &ConfigValidationError{
			Field:  "elasticsearch",
			Value:  nil,
			Reason: "elasticsearch configuration is required",
		}
	}

	if len(cfg.Addresses) == 0 {
		return &ConfigValidationError{
			Field:  "elasticsearch.addresses",
			Value:  cfg.Addresses,
			Reason: "elasticsearch addresses cannot be empty",
		}
	}

	if cfg.IndexName == "" {
		return &ConfigValidationError{
			Field:  "elasticsearch.index_name",
			Value:  cfg.IndexName,
			Reason: "elasticsearch index name cannot be empty",
		}
	}

	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		return &ConfigValidationError{
			Field:  "elasticsearch.api_key",
			Value:  cfg.APIKey,
			Reason: "either API key or username/password must be provided",
		}
	}

	return nil
}

// validateSources validates the source configurations
func validateSources(sources []Source) error {
	if len(sources) == 0 {
		return &ConfigValidationError{
			Field:  "sources",
			Value:  sources,
			Reason: "at least one source must be configured",
		}
	}

	for i, source := range sources {
		if source.Name == "" {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].name", i),
				Value:  source.Name,
				Reason: "source name cannot be empty",
			}
		}

		if source.URL == "" {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].url", i),
				Value:  source.URL,
				Reason: "source URL cannot be empty",
			}
		}

		if source.RateLimit < time.Second {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].rate_limit", i),
				Value:  source.RateLimit,
				Reason: "source rate limit must be at least 1 second",
			}
		}

		if source.MaxDepth < 1 {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].max_depth", i),
				Value:  source.MaxDepth,
				Reason: "source max depth must be greater than 0",
			}
		}

		if len(source.AllowedDomains) == 0 {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].allowed_domains", i),
				Value:  source.AllowedDomains,
				Reason: "at least one allowed domain must be configured",
			}
		}

		if len(source.StartURLs) == 0 {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].start_urls", i),
				Value:  source.StartURLs,
				Reason: "at least one start URL must be configured",
			}
		}
	}

	return nil
}
