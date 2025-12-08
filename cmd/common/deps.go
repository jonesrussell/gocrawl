// Package common provides shared utilities for command implementations.
package common

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// CommandDeps holds common dependencies for all commands.
// Use this instead of context.Value for type-safe dependency injection.
type CommandDeps struct {
	Logger logger.Interface
	Config config.Interface
}

// Validate ensures all required dependencies are present.
func (d CommandDeps) Validate() error {
	if d.Logger == nil {
		return ErrLoggerRequired
	}
	if d.Config == nil {
		return ErrConfigRequired
	}
	return nil
}
