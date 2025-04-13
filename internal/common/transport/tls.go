// Package transport provides common transport configuration for HTTP clients.
package transport

import (
	"crypto/tls"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config/crawler"
)

// NewTLSConfig creates a new TLS configuration from the given crawler config.
func NewTLSConfig(cfg *crawler.Config) (*tls.Config, error) {
	if err := cfg.TLS.Validate(); err != nil {
		return nil, fmt.Errorf("invalid TLS configuration: %w", err)
	}

	return &tls.Config{
		// InsecureSkipVerify is used here because some sources may have self-signed certificates
		// or other TLS issues. This is a trade-off between security and functionality.
		// In production, this should be properly configured based on the source's requirements.
		// WARNING: Setting this to true disables certificate verification and makes HTTPS connections
		// vulnerable to man-in-the-middle attacks. Only use this in development or with trusted sources.
		// #nosec G402 - This is intentional for development/testing purposes
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		// Set minimum TLS version to 1.2 for better security
		MinVersion: cfg.TLS.MinVersion,
		// Set maximum TLS version
		MaxVersion: cfg.TLS.MaxVersion,
		// Prefer server cipher suites
		PreferServerCipherSuites: cfg.TLS.PreferServerCipherSuites,
	}, nil
}
