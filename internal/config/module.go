// Package config provides configuration management for the GoCrawl application.
package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration module.
var Module = fx.Options(
	fx.Provide(
		func(cfg *Config) Interface {
			return cfg
		},
	),
	fx.Provide(
		func(path string) (*Config, error) {
			return LoadConfig(path)
		},
	),
)
