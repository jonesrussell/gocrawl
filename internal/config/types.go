// Package config provides configuration management for the application.
package config

import (
	"time"
)

// TLSConfig represents the TLS configuration.
type TLSConfig struct {
	// Enabled is whether TLS is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// CertFile is the certificate file path.
	CertFile string `json:"cert_file" yaml:"cert_file"`

	// KeyFile is the key file path.
	KeyFile string `json:"key_file" yaml:"key_file"`
}

// CORSConfig represents the CORS configuration.
type CORSConfig struct {
	// Enabled is whether CORS is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// AllowedOrigins is the list of allowed origins.
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`

	// AllowedMethods is the list of allowed methods.
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`

	// AllowedHeaders is the list of allowed headers.
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`

	// ExposedHeaders is the list of exposed headers.
	ExposedHeaders []string `json:"exposed_headers" yaml:"exposed_headers"`

	// AllowCredentials is whether to allow credentials.
	AllowCredentials bool `json:"allow_credentials" yaml:"allow_credentials"`

	// MaxAge is the maximum age of the preflight request.
	MaxAge time.Duration `json:"max_age" yaml:"max_age"`
}

// RateLimitConfig represents the rate limit configuration.
type RateLimitConfig struct {
	// Enabled is whether rate limiting is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RequestsPerSecond is the number of requests per second.
	RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`

	// BurstSize is the burst size.
	BurstSize int `json:"burst_size" yaml:"burst_size"`
}
