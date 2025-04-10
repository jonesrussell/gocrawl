// Package config provides configuration management for the GoCrawl application.
package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration module for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		LoadConfig,
	),
)
