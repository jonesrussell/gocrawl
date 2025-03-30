// Package config provides configuration functionality for the application.
package config

import (
	"go.uber.org/fx"
)

// Module provides the config module for dependency injection.
var Module = fx.Module("config",
	fx.Provide(
		// Provide the config with configuration
		fx.Annotate(
			NewConfig,
			fx.As(new(Interface)),
		),
	),
)

// NewConfig creates a new config instance with the given configuration.
func NewConfig(p Params) (Interface, error) {
	// For now, return a no-op config
	// In the future, this will load configuration from environment variables
	// and configuration files
	return NewNoOp(), nil
}
