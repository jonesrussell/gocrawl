// Package common provides shared utilities for command implementations.
package common

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// CommandContextKey is the key used to store the CommandContext in the context
var CommandContextKey contextKey = "commandContext"

// CommandContext holds typed dependencies for commands, replacing the context.Value anti-pattern
type CommandContext struct {
	Logger logger.Interface
	Config config.Interface
}

// LoggerKey is the key used to store the logger in the context (deprecated: use CommandContext)
var LoggerKey contextKey = "logger"

// ConfigKey is the context key for the configuration (deprecated: use CommandContext)
var ConfigKey contextKey = "config"
