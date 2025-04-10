// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
	stdlog "log"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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

	stdlog.Printf("Starting config validation, environment: %s", cfg.App.Environment)

	// Validate environment first
	if cfg.App.Environment == "" {
		stdlog.Printf("Environment validation failed: environment cannot be empty")
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
		stdlog.Printf("Environment validation failed: %s", cfg.App.Environment)
		return &ValidationError{
			Field:  "app.environment",
			Value:  cfg.App.Environment,
			Reason: "invalid environment",
		}
	}
	stdlog.Printf("Environment validation passed")

	// Validate log config
	if err := validateLogConfig(cfg.Logger); err != nil {
		stdlog.Printf("Log config validation failed: %v", err)
		return err
	}
	stdlog.Printf("Log config validation passed")

	// Validate crawler config
	if err := validateCrawlerConfig(cfg.Crawler); err != nil {
		stdlog.Printf("Crawler config validation failed: %v", err)
		return err
	}
	stdlog.Printf("Crawler config validation passed")

	// Validate Elasticsearch config
	if err := validateElasticsearchConfig(cfg.Elasticsearch); err != nil {
		stdlog.Printf("Elasticsearch config validation failed: %v", err)
		return err
	}
	stdlog.Printf("Elasticsearch config validation passed")

	// Validate app config last
	if err := validateAppConfig(cfg.App); err != nil {
		stdlog.Printf("App config validation failed: %v", err)
		return &ValidationError{
			Field:  "app",
			Value:  cfg.App,
			Reason: err.Error(),
		}
	}
	stdlog.Printf("App config validation passed")

	// Validate sources last
	if err := validateSources(cfg.Sources); err != nil {
		stdlog.Printf("Sources validation failed: %v", err)
		return err
	}
	stdlog.Printf("Sources validation passed")

	return nil
}

// ValidateEnvironment validates the environment setting.
func ValidateEnvironment(env string) error {
	switch env {
	case "development", "staging", "production":
		return nil
	default:
		return fmt.Errorf("invalid environment: %s", env)
	}
}

// ValidateLogLevel validates the log level setting.
func ValidateLogLevel(level string) error {
	switch level {
	case "debug", "info", "warn", "error":
		return nil
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
}

// ValidateRateLimit validates the rate limit setting.
func ValidateRateLimit(limit time.Duration) error {
	if limit < 0 {
		return errors.New("rate limit must be non-negative")
	}
	return nil
}

// ValidateMaxDepth validates the maximum depth setting.
func ValidateMaxDepth(depth int) error {
	if depth < 0 {
		return errors.New("max depth must be non-negative")
	}
	return nil
}

// ValidateParallelism validates the parallelism setting.
func ValidateParallelism(parallelism int) error {
	if parallelism < 1 {
		return errors.New("parallelism must be positive")
	}
	return nil
}

// ValidateSource validates a source configuration.
func ValidateSource(source *types.Source) error {
	if source == nil {
		return errors.New("source is required")
	}
	return source.Validate()
}

// ValidateSources validates a list of source configurations.
func ValidateSources(sources []types.Source) error {
	if len(sources) == 0 {
		return errors.New("at least one source is required")
	}
	for i := range sources {
		if err := sources[i].Validate(); err != nil {
			return fmt.Errorf("source[%d]: %w", i, err)
		}
	}
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
func validateLogConfig(cfg *log.Config) error {
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

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[strings.ToLower(cfg.Level)] {
		return &ValidationError{
			Field:  "log.level",
			Value:  cfg.Level,
			Reason: "invalid log level",
		}
	}
	return nil
}

// validateCrawlerConfig validates the crawler configuration
func validateCrawlerConfig(cfg *crawler.Config) error {
	if cfg == nil {
		return &ValidationError{
			Field:  "crawler",
			Value:  nil,
			Reason: "crawler configuration is required",
		}
	}
	return cfg.Validate()
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg *elasticsearch.Config) error {
	if cfg == nil {
		return &ValidationError{
			Field:  "elasticsearch",
			Value:  nil,
			Reason: "elasticsearch configuration is required",
		}
	}
	return cfg.Validate()
}

// validateSources validates a list of source configurations
func validateSources(sources []types.Source) error {
	if len(sources) == 0 {
		return &ValidationError{
			Field:  "sources",
			Value:  nil,
			Reason: "at least one source is required",
		}
	}

	for i := range sources {
		if err := sources[i].Validate(); err != nil {
			return fmt.Errorf("source[%d]: %w", i, err)
		}
	}
	return nil
}
