package config

import (
	"go.uber.org/fx"
)

// Module provides the config module and its dependencies
var Module = fx.Options(
	fx.Provide(
		NewConfig,        // Provide the NewConfig function to return *Config
		NewHTTPTransport, // Ensure this is also provided if needed
	),
)
