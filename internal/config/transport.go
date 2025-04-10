// Package config provides configuration management for the GoCrawl application.
package config

import (
	"net/http"
	"time"

	"go.uber.org/fx"
)

// TransportConfig represents the HTTP transport configuration.
type TransportConfig struct {
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout time.Duration `yaml:"tls_handshake_timeout"`
}

// NewTransportConfig creates a new transport configuration with default values.
func NewTransportConfig() *TransportConfig {
	return &TransportConfig{
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
		TLSHandshakeTimeout: DefaultTLSHandshakeTimeout,
	}
}

// TransportModule provides the HTTP transport configuration
var TransportModule = fx.Module("transport",
	fx.Provide(
		fx.Annotate(
			NewHTTPTransport,
			fx.As(new(http.RoundTripper)),
		),
	),
)

// NewHTTPTransport creates a new HTTP transport with default settings.
func NewHTTPTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:          DefaultMaxIdleConns,
		MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		TLSHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
