// Package config provides configuration management for the GoCrawl application.
package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration module.
var Module = fx.Options(
	fx.Provide(New),
	fx.Provide(LoadConfig),
)
