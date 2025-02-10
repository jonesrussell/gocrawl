package config

import (
	"net/http"

	"go.uber.org/fx"
)

// Module provides the configuration as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("config",
	fx.Provide(
		func() *Config {
			cfg, err := NewConfig(http.DefaultTransport)
			if err != nil {
				panic(err) // Or handle error appropriately
			}
			return cfg
		},
	),
)
