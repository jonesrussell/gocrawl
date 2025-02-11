package config

import (
	"net/http"

	"go.uber.org/fx"
)

// Module provides the config as an Fx module
var Module = fx.Module("config",
	fx.Provide(
		func(transport http.RoundTripper) (*Config, error) {
			// Use NewConfig to load configuration
			return NewConfig(transport)
		},
	),
)
