// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ValidateConfig orchestrates the validation of the entire configuration.
func ValidateConfig(cfg Interface) error {
	// Check app environment first
	appConfig := cfg.GetAppConfig()
	if appConfig.Environment == "production" {
		esConfig := cfg.GetElasticsearchConfig()
		if esConfig.APIKey == "" {
			return errors.New("API key is required in production")
		}
	}

	// Check server security based on command
	serverConfig := cfg.GetServerConfig()
	if serverConfig.Security.Enabled && cfg.GetCommand() == "httpd" {
		if serverConfig.Security.APIKey == "" {
			return errors.New("server API key is required when security is enabled for httpd command")
		}
		if _, err := uuid.Parse(serverConfig.Security.APIKey); err != nil {
			return fmt.Errorf("invalid server API key format: %w", err)
		}
	}

	// Validate individual components
	if err := validateAppConfig(appConfig); err != nil {
		return err
	}
	if err := validateElasticsearchConfig(cfg.GetElasticsearchConfig()); err != nil {
		return err
	}
	if err := validateLogConfig(cfg.GetLogConfig()); err != nil {
		return err
	}
	if err := validateCrawlerConfig(cfg.GetCrawlerConfig()); err != nil {
		return err
	}
	return nil
}

// validateAppConfig validates the app-specific configuration.
func validateAppConfig(cfg *AppConfig) error {
	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}
	if !validEnvironments[cfg.Environment] {
		return fmt.Errorf("invalid app environment: %s", cfg.Environment)
	}
	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration.
func validateElasticsearchConfig(cfg *ElasticsearchConfig) error {
	if len(cfg.Addresses) == 0 {
		return errors.New("elasticsearch addresses are required")
	}
	return nil
}

// validateLogConfig validates the log-related configuration.
func validateLogConfig(cfg *LogConfig) error {
	if !isValidLogLevel(cfg.Level) {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	return nil
}

// isValidLogLevel checks if the given log level is valid.
func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	return validLevels[level]
}

// validateCrawlerConfig validates the crawler-specific configuration.
func validateCrawlerConfig(cfg *CrawlerConfig) error {
	if cfg.MaxDepth < 1 {
		return errors.New("crawler max depth must be greater than 0")
	}
	if cfg.Parallelism < 1 {
		return errors.New("crawler parallelism must be greater than 0")
	}
	return nil
}
