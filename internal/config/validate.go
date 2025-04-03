// Package config provides configuration management for the GoCrawl application.
package config

import (
	"fmt"
	"strings"
	"time"
)

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	if err := validateAppConfig(cfg.App); err != nil {
		return err
	}
	if err := validateCrawlerConfig(cfg.Crawler); err != nil {
		return err
	}
	if err := validateElasticsearchConfig(cfg.Elasticsearch); err != nil {
		return err
	}
	if err := validateServerConfig(cfg.Server); err != nil {
		return err
	}
	if err := validateSources(cfg.Sources); err != nil {
		return err
	}
	return nil
}

// validateAppConfig validates the application configuration
func validateAppConfig(cfg AppConfig) error {
	if cfg.Environment == "" {
		return &ConfigValidationError{
			Field:  "app.environment",
			Value:  cfg.Environment,
			Reason: "environment cannot be empty",
		}
	}
	// Validate environment value
	validEnvs := []string{envDevelopment, envStaging, envProduction}
	isValidEnv := false
	for _, env := range validEnvs {
		if cfg.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return &ConfigValidationError{
			Field:  "app.environment",
			Value:  cfg.Environment,
			Reason: fmt.Sprintf("environment must be one of: %v", validEnvs),
		}
	}
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

// validateCrawlerConfig validates the crawler configuration
func validateCrawlerConfig(cfg CrawlerConfig) error {
	if cfg.BaseURL == "" {
		return &ConfigValidationError{
			Field:  "crawler.base_url",
			Value:  cfg.BaseURL,
			Reason: "base URL cannot be empty",
		}
	}
	if cfg.MaxDepth < 1 {
		return &ConfigValidationError{
			Field:  "crawler.max_depth",
			Value:  cfg.MaxDepth,
			Reason: "max depth must be greater than 0",
		}
	}
	if cfg.RateLimit < time.Second {
		return &ConfigValidationError{
			Field:  "crawler.rate_limit",
			Value:  cfg.RateLimit,
			Reason: "rate limit must be at least 1 second",
		}
	}
	if cfg.Parallelism < 1 {
		return &ConfigValidationError{
			Field:  "crawler.parallelism",
			Value:  cfg.Parallelism,
			Reason: "parallelism must be greater than 0",
		}
	}
	if cfg.SourceFile == "" {
		return &ConfigValidationError{
			Field:  "crawler.source_file",
			Value:  cfg.SourceFile,
			Reason: "source file cannot be empty",
		}
	}
	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg ElasticsearchConfig) error {
	if len(cfg.Addresses) == 0 {
		return &ConfigValidationError{
			Field:  "elasticsearch.addresses",
			Value:  cfg.Addresses,
			Reason: "at least one Elasticsearch address must be provided",
		}
	}

	// Validate each address
	for _, addr := range cfg.Addresses {
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			return &ConfigValidationError{
				Field:  "elasticsearch.addresses",
				Value:  addr,
				Reason: "invalid Elasticsearch address",
			}
		}
	}

	// Validate credentials
	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		return &ConfigValidationError{
			Field:  "elasticsearch.credentials",
			Value:  "missing",
			Reason: "either username/password or api_key must be provided",
		}
	}

	if cfg.IndexName == "" {
		return &ConfigValidationError{
			Field:  "elasticsearch.index_name",
			Value:  cfg.IndexName,
			Reason: "index name cannot be empty",
		}
	}

	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" {
			return &ConfigValidationError{
				Field:  "elasticsearch.tls.certificate",
				Value:  cfg.TLS.CertFile,
				Reason: "certificate path cannot be empty when TLS is enabled",
			}
		}
		if cfg.TLS.KeyFile == "" {
			return &ConfigValidationError{
				Field:  "elasticsearch.tls.key",
				Value:  cfg.TLS.KeyFile,
				Reason: "key path cannot be empty when TLS is enabled",
			}
		}
	}

	// Validate retry configuration if enabled
	if cfg.Retry.Enabled {
		if cfg.Retry.MaxRetries < 0 {
			return &ConfigValidationError{
				Field:  "elasticsearch.retry.max_retries",
				Value:  cfg.Retry.MaxRetries,
				Reason: "retry values must be non-negative",
			}
		}
		if cfg.Retry.InitialWait < time.Second {
			return &ConfigValidationError{
				Field:  "elasticsearch.retry.initial_wait",
				Value:  cfg.Retry.InitialWait,
				Reason: "initial wait must be at least 1 second",
			}
		}
		if cfg.Retry.MaxWait < cfg.Retry.InitialWait {
			return &ConfigValidationError{
				Field:  "elasticsearch.retry.max_wait",
				Value:  cfg.Retry.MaxWait,
				Reason: "max wait must be greater than or equal to initial wait",
			}
		}
	}

	return nil
}

// validateServerConfig validates the server configuration
func validateServerConfig(cfg ServerConfig) error {
	if cfg.Address == "" {
		return &ConfigValidationError{
			Field:  "server.address",
			Value:  cfg.Address,
			Reason: "server address cannot be empty",
		}
	}
	if cfg.ReadTimeout < time.Second {
		return &ConfigValidationError{
			Field:  "server.read_timeout",
			Value:  cfg.ReadTimeout,
			Reason: "read timeout must be at least 1 second",
		}
	}
	if cfg.WriteTimeout < time.Second {
		return &ConfigValidationError{
			Field:  "server.write_timeout",
			Value:  cfg.WriteTimeout,
			Reason: "write timeout must be at least 1 second",
		}
	}
	if cfg.IdleTimeout < time.Second {
		return &ConfigValidationError{
			Field:  "server.idle_timeout",
			Value:  cfg.IdleTimeout,
			Reason: "idle timeout must be at least 1 second",
		}
	}
	if cfg.Security.Enabled {
		if err := validateServerSecurity(cfg.Security); err != nil {
			return err
		}
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
	if security.APIKey == "" {
		return &ConfigValidationError{
			Field:  "server.security.api_key",
			Value:  security.APIKey,
			Reason: "API key cannot be empty when security is enabled",
		}
	}
	if security.TLS.Enabled {
		if err := validateServerTLS(security.TLS); err != nil {
			return err
		}
	}
	return nil
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
		return &ConfigValidationError{
			Field:  "sources",
			Value:  sources,
			Reason: "at least one source must be configured",
		}
	}
	for i := range sources {
		source := &sources[i]
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
		if source.MaxDepth < 1 {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].max_depth", i),
				Value:  source.MaxDepth,
				Reason: "source max depth must be greater than 0",
			}
		}
		if source.RateLimit < time.Second {
			return &ConfigValidationError{
				Field:  fmt.Sprintf("sources[%d].rate_limit", i),
				Value:  source.RateLimit,
				Reason: "source rate limit must be at least 1 second",
			}
		}
	}
	return nil
}
