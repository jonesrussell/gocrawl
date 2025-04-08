// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"go.uber.org/zap"
)

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
		return &ValidationError{
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
		return &ValidationError{
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
		return &ValidationError{
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
func validateAppConfig(cfg *app.Config) error {
	if cfg.Name == "" {
		return &ValidationError{
			Field:  "app.name",
			Value:  cfg.Name,
			Reason: "name cannot be empty",
		}
	}
	if cfg.Version == "" {
		return &ValidationError{
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
		return &ValidationError{
			Field:  "log",
			Value:  nil,
			Reason: "log configuration is required",
		}
	}

	if cfg.Level == "" {
		cfg.Level = "info" // Default to info if not set
	}

	if !ValidLogLevels[strings.ToLower(cfg.Level)] {
		return &ValidationError{
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
		return &ValidationError{
			Field:  "crawler",
			Value:  nil,
			Reason: "crawler configuration is required",
		}
	}

	if cfg.BaseURL == "" {
		return &ValidationError{
			Field:  "crawler.base_url",
			Value:  cfg.BaseURL,
			Reason: "crawler base URL cannot be empty",
		}
	}

	if cfg.MaxDepth < 1 {
		return &ValidationError{
			Field:  "crawler.max_depth",
			Value:  cfg.MaxDepth,
			Reason: "crawler max depth must be greater than 0",
		}
	}

	if cfg.RateLimit < time.Second {
		return &ValidationError{
			Field:  "crawler.rate_limit",
			Value:  cfg.RateLimit,
			Reason: "crawler rate limit must be at least 1 second",
		}
	}

	if cfg.RandomDelay < 0 {
		return &ValidationError{
			Field:  "crawler.random_delay",
			Value:  cfg.RandomDelay,
			Reason: "crawler random delay must be non-negative",
		}
	}

	if cfg.Parallelism < 1 {
		return &ValidationError{
			Field:  "crawler.parallelism",
			Value:  cfg.Parallelism,
			Reason: "crawler parallelism must be greater than 0",
		}
	}

	if cfg.SourceFile == "" {
		return &ValidationError{
			Field:  "crawler.source_file",
			Value:  cfg.SourceFile,
			Reason: "crawler source file is required",
		}
	}

	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg *ElasticsearchConfig) error {
	if cfg == nil {
		return &ValidationError{
			Field:  "elasticsearch",
			Value:  nil,
			Reason: "elasticsearch configuration is required",
		}
	}

	if len(cfg.Addresses) == 0 {
		return &ValidationError{
			Field:  "elasticsearch.addresses",
			Value:  cfg.Addresses,
			Reason: "at least one Elasticsearch address is required",
		}
	}

	if cfg.IndexName == "" {
		return &ValidationError{
			Field:  "elasticsearch.index_name",
			Value:  cfg.IndexName,
			Reason: "elasticsearch index name is required",
		}
	}

	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		return &ValidationError{
			Field:  "elasticsearch.auth",
			Value:  nil,
			Reason: "either API key or username/password is required",
		}
	}

	return nil
}

// validateSources validates the source configuration
func validateSources(sources []Source) error {
	if len(sources) == 0 {
		return &ValidationError{
			Field:  "sources",
			Value:  sources,
			Reason: "at least one source is required",
		}
	}

	for i, source := range sources {
		if source.Name == "" {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].name", i),
				Value:  source.Name,
				Reason: "source name cannot be empty",
			}
		}

		if source.URL == "" {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].url", i),
				Value:  source.URL,
				Reason: "source URL cannot be empty",
			}
		}

		if len(source.AllowedDomains) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].allowed_domains", i),
				Value:  source.AllowedDomains,
				Reason: "at least one allowed domain is required",
			}
		}

		if len(source.StartURLs) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].start_urls", i),
				Value:  source.StartURLs,
				Reason: "at least one start URL is required",
			}
		}

		if source.RateLimit < time.Second {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].rate_limit", i),
				Value:  source.RateLimit,
				Reason: "rate limit must be at least 1 second",
			}
		}

		if source.MaxDepth < 1 {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].max_depth", i),
				Value:  source.MaxDepth,
				Reason: "max depth must be greater than 0",
			}
		}

		if len(source.Selectors) == 0 {
			return &ValidationError{
				Field:  fmt.Sprintf("sources[%d].selectors", i),
				Value:  source.Selectors,
				Reason: "at least one selector is required",
			}
		}
	}

	return nil
}
