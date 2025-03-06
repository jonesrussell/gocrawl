package config

import (
	"go.uber.org/fx"
)

// Module provides the config module and its dependencies
var Module = fx.Options(
	fx.Provide(
		New,              // Provide the New function to return *Config
		NewHTTPTransport, // Ensure this is also provided if needed
	),
)
