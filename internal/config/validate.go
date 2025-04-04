// Package config provides configuration management for the GoCrawl application.
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	// Validate app config first
	if err := validateAppConfig(&cfg.App); err != nil {
		return fmt.Errorf("app validation failed: %w", err)
	}

	// Validate log config
	if err := validateLogConfig(&cfg.Log); err != nil {
		return fmt.Errorf("log validation failed: %w", err)
	}

	// Validate Elasticsearch config
	if err := validateElasticsearchConfig(&cfg.Elasticsearch); err != nil {
		return fmt.Errorf("Elasticsearch validation failed: %w", err)
	}

	// Validate crawler config
	if err := validateCrawlerConfig(&cfg.Crawler); err != nil {
		return fmt.Errorf("crawler validation failed: %w", err)
	}

	// Validate server config
	if err := validateServerConfig(&cfg.Server); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	// Skip sources validation in test environment
	if cfg.App.Environment != "test" {
		if err := validateSources(cfg.Sources); err != nil {
			return err
		}
	}
	return nil
}

// validateAppConfig validates the application configuration
func validateAppConfig(cfg *AppConfig) error {
	if cfg.Environment == "" {
		return errors.New("environment cannot be empty")
	}
	// Validate environment value
	validEnvs := []string{envDevelopment, envStaging, envProduction, envTest}
	isValidEnv := false
	for _, env := range validEnvs {
		if cfg.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return fmt.Errorf("invalid environment: %s", cfg.Environment)
	}
	if cfg.Name == "" {
		return errors.New("name cannot be empty")
	}
	if cfg.Version == "" {
		return errors.New("version cannot be empty")
	}
	return nil
}

// validateLogConfig validates the log configuration
func validateLogConfig(cfg *LogConfig) error {
	if cfg.Level == "" {
		cfg.Level = "info" // Default to info if not set
	}
	validLevels := []string{"debug", "info", "warn", "error"}
	isValidLevel := false
	for _, level := range validLevels {
		if strings.ToLower(cfg.Level) == level {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	return nil
}

// validateCrawlerConfig validates the crawler configuration
func validateCrawlerConfig(cfg *CrawlerConfig) error {
	if cfg.BaseURL == "" {
		return errors.New("crawler base URL cannot be empty")
	}

	if cfg.MaxDepth < 1 {
		return errors.New("crawler max depth must be greater than 0")
	}

	if cfg.RateLimit < time.Second {
		return errors.New("crawler rate limit must be at least 1 second")
	}

	if cfg.Parallelism < 1 {
		return errors.New("crawler parallelism must be greater than 0")
	}

	if cfg.SourceFile == "" {
		return errors.New("crawler source file cannot be empty")
	}

	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg *ElasticsearchConfig) error {
	if len(cfg.Addresses) == 0 {
		return errors.New("elasticsearch addresses cannot be empty")
	}

	if cfg.IndexName == "" {
		return errors.New("elasticsearch index name cannot be empty")
	}

	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		return errors.New("elasticsearch API key cannot be empty")
	}

	// Validate API key format if provided
	if cfg.APIKey != "" && !strings.Contains(cfg.APIKey, ":") {
		return errors.New("elasticsearch API key must be in the format 'id:api_key'")
	}

	// Validate TLS configuration
	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" {
			return errors.New("TLS certificate file is required when TLS is enabled")
		}
	}

	return nil
}

// validateServerConfig validates the server configuration
func validateServerConfig(cfg *ServerConfig) error {
	if cfg.Address == "" {
		return &ConfigValidationError{
			Field:  "server.address",
			Value:  cfg.Address,
			Reason: "address cannot be empty",
		}
	}

	if err := validateServerSecurity(cfg.Security); err != nil {
		return err
	}

	return nil
}

// validateServerSecurity validates the server security configuration
func validateServerSecurity(security struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key"`
	RateLimit int    `yaml:"rate_limit"`
	CORS      struct {
		Enabled        bool     `yaml:"enabled"`
		AllowedOrigins []string `yaml:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers"`
		MaxAge         int      `yaml:"max_age"`
	} `yaml:"cors"`
	TLS TLSConfig `yaml:"tls"`
}) error {
	if security.Enabled {
		if security.APIKey == "" {
			return errors.New("server security is enabled but no API key is provided")
		}
		if !isValidAPIKey(security.APIKey) {
			return errors.New("invalid API key format")
		}
	}

	if security.RateLimit < 0 {
		return &ConfigValidationError{
			Field:  "server.security.rate_limit",
			Value:  security.RateLimit,
			Reason: "rate limit must be non-negative",
		}
	}

	if security.CORS.Enabled {
		if len(security.CORS.AllowedOrigins) == 0 {
			return &ConfigValidationError{
				Field:  "server.security.cors.allowed_origins",
				Value:  security.CORS.AllowedOrigins,
				Reason: "at least one allowed origin must be specified when CORS is enabled",
			}
		}
		if len(security.CORS.AllowedMethods) == 0 {
			return &ConfigValidationError{
				Field:  "server.security.cors.allowed_methods",
				Value:  security.CORS.AllowedMethods,
				Reason: "at least one allowed method must be specified when CORS is enabled",
			}
		}
		if security.CORS.MaxAge < 0 {
			return &ConfigValidationError{
				Field:  "server.security.cors.max_age",
				Value:  security.CORS.MaxAge,
				Reason: "max age must be non-negative",
			}
		}
	}

	if security.TLS.Enabled {
		if err := validateServerTLS(security.TLS); err != nil {
			return err
		}
	}

	return nil
}

// isValidAPIKey checks if the API key has a valid format
func isValidAPIKey(key string) bool {
	// API key must be at least 32 characters long and contain only alphanumeric characters
	if len(key) < 32 {
		return false
	}
	for _, c := range key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// validateServerTLS validates the server TLS configuration
func validateServerTLS(tls TLSConfig) error {
	if tls.CertFile == "" {
		return &ConfigValidationError{
			Field:  "server.security.tls.certificate",
			Value:  tls.CertFile,
			Reason: "certificate path cannot be empty when TLS is enabled",
		}
	}
	if tls.KeyFile == "" {
		return &ConfigValidationError{
			Field:  "server.security.tls.key",
			Value:  tls.KeyFile,
			Reason: "key path cannot be empty when TLS is enabled",
		}
	}
	return nil
}

// validateSources validates the source configurations
func validateSources(sources []Source) error {
	if len(sources) == 0 {
		return errors.New("at least one source must be configured")
	}
	for i := range sources {
		source := &sources[i]
		if source.Name == "" {
			return fmt.Errorf("source[%d] name cannot be empty", i)
		}
		if source.URL == "" {
			return fmt.Errorf("source[%d] URL cannot be empty", i)
		}
		if source.MaxDepth < 1 {
			return fmt.Errorf("source[%d] max depth must be greater than 0", i)
		}
		if source.RateLimit < time.Second {
			return fmt.Errorf("source[%d] rate limit must be at least 1 second", i)
		}
	}
	return nil
}
