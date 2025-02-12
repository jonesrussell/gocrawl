package config

import (
	"net/http"

	"go.uber.org/fx"
)

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// Module provides the configuration as a dependency
var Module = fx.Module("config",
	fx.Provide(
		NewHTTPTransport, // Ensure this is provided
		NewConfig,
	),
)
