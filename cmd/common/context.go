// Package common provides shared utilities for command implementations.
package common

// CommandContext and related context.Value keys are deprecated.
// This file is kept for backward compatibility during migration but should not be used.
//
// DEPRECATED: Use CommandDeps struct instead of context.Value for dependency injection.
//
// Old approach (deprecated):
//   ctx := context.WithValue(ctx, CommandContextKey, cmdCtx)
//   cmdCtx := ctx.Value(CommandContextKey).(*CommandContext)
//
// New approach (recommended):
//   deps := CommandDeps{Logger: log, Config: cfg}
//   // Pass deps directly to functions
//
// See cmd/common/deps.go for the new approach.

import (
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// contextKey is a custom type for context keys (deprecated)
type contextKey string

// CommandContextKey is deprecated. Use CommandDeps directly instead.
var CommandContextKey contextKey = "commandContext"

// CommandContext is deprecated. Use CommandDeps directly instead.
type CommandContext struct {
	Logger logger.Interface
	Config config.Interface
}

// LoggerKey is deprecated. Use CommandDeps.Logger instead.
var LoggerKey contextKey = "logger"

// ConfigKey is deprecated. Use CommandDeps.Config instead.
var ConfigKey contextKey = "config"
