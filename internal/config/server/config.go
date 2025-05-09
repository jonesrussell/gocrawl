// Package server provides server configuration types and functions.
package server

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// APIKeyParts is the number of parts in an API key (id:key)
const APIKeyParts = 2

// Config represents server-specific configuration settings.
type Config struct {
	// Host is the server host address
	Host string `yaml:"host"`
	// Port is the server port number
	Port int `yaml:"port"`
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `yaml:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `yaml:"write_timeout"`
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header
	MaxHeaderBytes int `yaml:"max_header_bytes"`
	// SecurityEnabled determines if security features are enabled
	SecurityEnabled bool `yaml:"security_enabled"`
	// APIKey is the API key used for authentication
	APIKey string `yaml:"api_key"`
	// Address is the address to listen on (e.g., ":8080")
	Address string `yaml:"address"`
	// TLS contains TLS configuration
	TLS struct {
		// Enabled indicates whether TLS is enabled
		Enabled bool `yaml:"enabled"`
		// CertFile is the path to the certificate file
		CertFile string `yaml:"cert_file"`
		// KeyFile is the path to the key file
		KeyFile string `yaml:"key_file"`
	} `yaml:"tls"`
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.SecurityEnabled {
		if c.APIKey == "" {
			return errors.New("server security is enabled but no API key is provided")
		}

		// Validate API key format
		parts := strings.Split(c.APIKey, ":")
		if len(parts) != APIKeyParts {
			return fmt.Errorf("invalid API key format: expected 'id:key' but got %q", c.APIKey)
		}

		// Validate API key parts
		id, key := parts[0], parts[1]
		if id == "" {
			return errors.New("API key ID cannot be empty")
		}
		if key == "" {
			return errors.New("API key value cannot be empty")
		}
	}

	return nil
}

// New creates a new server configuration with the given options.
func New(opts ...Option) *Config {
	cfg := &Config{}

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

// NewConfig creates a new Config instance with default values.
func NewConfig() *Config {
	return &Config{
		TLS: struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		}{
			Enabled: false,
		},
	}
}
