package config

import (
	"go.uber.org/fx"
)

// Module provides the config module and its dependencies
var Module = fx.Module("config",
	fx.Provide(
		NewConfig,        // Function to create a new config instance
		NewHTTPTransport, // Ensure this is provided
	),
)
