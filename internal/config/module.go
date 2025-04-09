// Package config provides configuration management for the GoCrawl application.
package config

import "go.uber.org/fx"

// Module provides the configuration package's dependencies.
var Module = fx.Module("config",
	fx.Provide(
		NewConfig,
	),
)
