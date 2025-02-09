package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration as an Fx module
var Module = fx.Module("config",
	fx.Provide(
		LoadConfig, // Ensure LoadConfig is provided
	),
)
