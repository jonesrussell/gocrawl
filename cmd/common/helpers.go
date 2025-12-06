// Package common provides shared utilities for command implementations.
package common

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// GetCommandContext retrieves the CommandContext from the context.
// Returns an error if the context is not found or has an invalid type.
func GetCommandContext(ctx context.Context) (*CommandContext, error) {
	ctxValue := ctx.Value(CommandContextKey)
	if ctxValue == nil {
		return nil, errors.New("command context not found in context")
	}

	cmdCtx, ok := ctxValue.(*CommandContext)
	if !ok {
		return nil, errors.New("command context has invalid type")
	}

	return cmdCtx, nil
}

// GetDependencies retrieves logger and config from context.
// This is a convenience function that extracts dependencies from CommandContext.
// Returns an error if dependencies are not found or have invalid types.
func GetDependencies(ctx context.Context) (logger.Interface, config.Interface, error) {
	cmdCtx, err := GetCommandContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if cmdCtx.Logger == nil {
		return nil, nil, errors.New("logger not found in command context")
	}

	if cmdCtx.Config == nil {
		return nil, nil, errors.New("config not found in command context")
	}

	return cmdCtx.Logger, cmdCtx.Config, nil
}
