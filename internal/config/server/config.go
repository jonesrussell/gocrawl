package server

import (
	"errors"
	"fmt"
	"strings"
)

// Default configuration values
const (
	DefaultSecurityEnabled = false
	DefaultAPIKey          = ""
)

// Config holds server-specific configuration settings.
type Config struct {
	// SecurityEnabled determines if security features are enabled
	SecurityEnabled bool `yaml:"security_enabled"`
	// APIKey is the API key used for authentication
	APIKey string `yaml:"api_key"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.SecurityEnabled {
		if c.APIKey == "" {
			return errors.New("server security is enabled but no API key is provided")
		}

		// API key should be in the format "id:api_key"
		parts := strings.Split(c.APIKey, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("server API key must be in the format 'id:api_key'")
		}
	}

	return nil
}

// New creates a new server configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{
		SecurityEnabled: DefaultSecurityEnabled,
		APIKey:          DefaultAPIKey,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// Option is a function that configures a server configuration.
type Option func(*Config)

// WithSecurityEnabled sets whether security features are enabled.
func WithSecurityEnabled(enabled bool) Option {
	return func(c *Config) {
		c.SecurityEnabled = enabled
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}
