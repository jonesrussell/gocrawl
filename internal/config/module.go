package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("config",
	fx.Provide(
		LoadConfig, // Ensure LoadConfig is provided
	),
)
