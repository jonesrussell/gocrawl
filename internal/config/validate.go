// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ValidateConfig orchestrates the validation of the entire configuration.
func ValidateConfig(cfg *Config) error {
	if err := validateAppConfig(cfg); err != nil {
		return err
	}
	if err := validateElasticsearchConfig(cfg); err != nil {
		return err
	}
	if err := validateLogConfig(cfg); err != nil {
		return err
	}
	if err := validateServerConfig(cfg); err != nil {
		return err
	}
	if err := validateCrawlerConfig(cfg); err != nil {
		return err
	}
	return nil
}

// validateAppConfig validates the app-specific configuration.
func validateAppConfig(cfg *Config) error {
	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}
	if !validEnvironments[cfg.App.Environment] {
		return fmt.Errorf("invalid app environment: %s", cfg.App.Environment)
	}
	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration.
func validateElasticsearchConfig(cfg *Config) error {
	if cfg.App.Environment == "production" && cfg.Elasticsearch.APIKey == "" {
		return errors.New("API key is required in production")
	}
	return nil
}

// validateLogConfig validates the log-related configuration.
func validateLogConfig(cfg *Config) error {
	if !isValidLogLevel(cfg.Log.Level) {
		return fmt.Errorf("invalid log level: %s", cfg.Log.Level)
	}
	return nil
}

// isValidLogLevel checks if the given log level is valid.
func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	return validLevels[level]
}

// validateServerConfig validates the server-related configuration.
func validateServerConfig(cfg *Config) error {
	if cfg.Server.Security.Enabled {
		if cfg.Server.Security.APIKey == "" {
			return errors.New("server API key is required when security is enabled")
		}
		if _, err := uuid.Parse(cfg.Server.Security.APIKey); err != nil {
			return fmt.Errorf("invalid server API key format: %w", err)
		}
	}
	return nil
}

// validateCrawlerConfig validates the crawler-specific configuration.
func validateCrawlerConfig(cfg *Config) error {
	if cfg.Crawler.MaxDepth < 1 {
		return errors.New("crawler max depth must be greater than 0")
	}
	if cfg.Crawler.Parallelism < 1 {
		return errors.New("crawler parallelism must be greater than 0")
	}
	return nil
}
