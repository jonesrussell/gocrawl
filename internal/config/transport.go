// Package config provides configuration management for the GoCrawl application.
package config

import (
	"net/http"
	"time"
)

// TransportConfig holds the configuration for HTTP transport.
type TransportConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     int
	TLSHandshakeTimeout int
}

// NewTransportConfig initializes a TransportConfig with default values.
func NewTransportConfig() *TransportConfig {
	return &TransportConfig{
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     int(DefaultIdleConnTimeout.Seconds()),
		TLSHandshakeTimeout: int(DefaultTLSHandshakeTimeout.Seconds()),
	}
}

// TransportModule provides the HTTP transport configuration.
var TransportModule = NewTransportConfig()

// NewHTTPTransport creates a new HTTP transport using the provided configuration.
func NewHTTPTransport(config *TransportConfig) *http.Transport {
	return &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(config.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout: time.Duration(config.TLSHandshakeTimeout) * time.Second,
	}
}
