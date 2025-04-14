// Package common provides shared utilities for command implementations.
package common

import (
	"context"
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// LoggerKey is the key used to store the logger in the context
var LoggerKey contextKey = "logger"

// ConfigKey is the context key for the configuration
var ConfigKey contextKey = "config"

// GetLoggerFromContext retrieves the logger from the context
func GetLoggerFromContext(ctx context.Context) (*zap.Logger, error) {
	if logger, ok := ctx.Value(LoggerKey).(*zap.Logger); ok {
		return logger, nil
	}
	return nil, errors.New("logger not found in context or invalid type")
}

// GetConfigFromContext retrieves the config from the context
func GetConfigFromContext(ctx context.Context) (*viper.Viper, error) {
	if config, ok := ctx.Value(ConfigKey).(*viper.Viper); ok {
		return config, nil
	}
	return nil, errors.New("config not found in context or invalid type")
}
