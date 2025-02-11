package config

import (
	"net/http"

	"go.uber.org/fx"
)

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// Module provides the config as an Fx module
var Module = fx.Module("config",
	fx.Provide(
		NewHTTPTransport, // Provide the HTTP transport
		func(transport http.RoundTripper) (*Config, error) {
			// Use NewConfig to load configuration
			return NewConfig(transport)
		},
	),
)
