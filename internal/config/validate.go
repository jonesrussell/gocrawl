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
	// Debug: Print configuration being validated
	fmt.Printf("Validating configuration:\n")
	fmt.Printf("  App: %+v\n", cfg.App)
	fmt.Printf("  Log: %+v\n", cfg.Log)
	fmt.Printf("  Elasticsearch: %+v\n", cfg.Elasticsearch)
	fmt.Printf("  Crawler: %+v\n", cfg.Crawler)
	fmt.Printf("  Server: %+v\n", cfg.Server)
	fmt.Printf("  Sources: %+v\n", cfg.Sources)

	// Validate specific configurations first
	if err := validateAppConfig(cfg.App); err != nil {
		fmt.Printf("App validation failed: %v\n", err)
		return err
	}
	if err := validateLogConfig(cfg.Log); err != nil {
		fmt.Printf("Log validation failed: %v\n", err)
		return err
	}
	if err := validateElasticsearchConfig(cfg.Elasticsearch); err != nil {
		fmt.Printf("Elasticsearch validation failed: %v\n", err)
		return err
	}
	if err := validateCrawlerConfig(cfg.Crawler); err != nil {
		fmt.Printf("Crawler validation failed: %v\n", err)
		return err
	}
	if err := validateServerConfig(cfg.Server); err != nil {
		fmt.Printf("Server validation failed: %v\n", err)
		return err
	}
	if err := validateSources(cfg.Sources); err != nil {
		fmt.Printf("Sources validation failed: %v\n", err)
		return err
	}
	return nil
}

// validateAppConfig validates the application configuration
func validateAppConfig(cfg AppConfig) error {
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
func validateLogConfig(cfg LogConfig) error {
	fmt.Printf("Validating log config: %+v\n", cfg)
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
		fmt.Printf("Invalid log level: %s\n", cfg.Level)
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	return nil
}

// validateCrawlerConfig validates the crawler configuration
func validateCrawlerConfig(cfg CrawlerConfig) error {
	fmt.Printf("DEBUG: Validating crawler config: %+v\n", cfg)

	if cfg.BaseURL == "" {
		fmt.Printf("DEBUG: BaseURL validation failed - empty\n")
		return errors.New("crawler base URL cannot be empty")
	}

	if cfg.MaxDepth < 1 {
		fmt.Printf("DEBUG: MaxDepth validation failed - value: %d\n", cfg.MaxDepth)
		return errors.New("crawler max depth must be greater than 0")
	}

	if cfg.RateLimit < time.Second {
		fmt.Printf("DEBUG: RateLimit validation failed - value: %v\n", cfg.RateLimit)
		return errors.New("crawler rate limit must be at least 1 second")
	}

	if cfg.Parallelism < 1 {
		fmt.Printf("DEBUG: Parallelism validation failed - value: %d\n", cfg.Parallelism)
		return errors.New("crawler parallelism must be greater than 0")
	}

	if cfg.SourceFile == "" {
		fmt.Printf("DEBUG: SourceFile validation failed - empty\n")
		return errors.New("crawler source file cannot be empty")
	}

	fmt.Printf("DEBUG: Crawler config validation passed\n")
	return nil
}

// validateElasticsearchConfig validates the Elasticsearch configuration
func validateElasticsearchConfig(cfg ElasticsearchConfig) error {
	fmt.Printf("Validating Elasticsearch config: %+v\n", cfg)

	// Validate addresses first
	if len(cfg.Addresses) == 0 {
		fmt.Printf("No Elasticsearch addresses provided\n")
		return &ConfigValidationError{
			Field:  "elasticsearch.addresses",
			Value:  cfg.Addresses,
			Reason: "at least one Elasticsearch address must be provided",
		}
	}

	// Validate each address
	for _, addr := range cfg.Addresses {
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			fmt.Printf("Invalid Elasticsearch address: %s\n", addr)
			return &ConfigValidationError{
				Field:  "elasticsearch.addresses",
				Value:  addr,
				Reason: "invalid Elasticsearch address",
			}
		}
	}

	// Validate TLS configuration if enabled
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

	// Validate credentials
	if cfg.APIKey == "" && (cfg.Username == "" || cfg.Password == "") {
		fmt.Printf("Missing Elasticsearch credentials\n")
		return &ConfigValidationError{
			Field:  "elasticsearch.credentials",
			Value:  "missing",
			Reason: "either username/password or api_key must be provided",
		}
	}

	// Validate index name
	if cfg.IndexName == "" {
		return &ConfigValidationError{
			Field:  "elasticsearch.index_name",
			Value:  cfg.IndexName,
			Reason: "index name cannot be empty",
		}
	}

	// Validate retry configuration if enabled
	if cfg.Retry.Enabled {
		if cfg.Retry.MaxRetries < 0 {
			return &ConfigValidationError{
				Field:  "elasticsearch.retry.max_retries",
				Value:  cfg.Retry.MaxRetries,
				Reason: "max retries must be greater than or equal to 0",
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
				Reason: "max wait must be greater than initial wait",
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
