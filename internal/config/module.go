package config

import (
	"net/http"

	"go.uber.org/fx"
)

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport() http.RoundTripper {
	return http.DefaultTransport
}

// Module provides the configuration module and its dependencies
var Module = fx.Module("config",
	fx.Provide(
		NewHTTPTransport, // Ensure this is provided
		NewConfig,        // Function to create a new config instance
	),
)
