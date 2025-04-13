// Package common provides shared utilities for command implementations.
package common

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/viper"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	// LoggerKey is the key used to store the logger in the context
	LoggerKey contextKey = "logger"
	// ConfigKey is the context key for the configuration
	ConfigKey contextKey = "config"
)

// GetLoggerFromContext returns the logger from the context
func GetLoggerFromContext(ctx context.Context) logger.Interface {
	if logger, ok := ctx.Value(LoggerKey).(logger.Interface); ok {
		return logger
	}
	return nil
}

// GetConfigFromContext returns the configuration from the context
func GetConfigFromContext(ctx context.Context) *viper.Viper {
	if config, ok := ctx.Value(ConfigKey).(*viper.Viper); ok {
		return config
	}
	return nil
}
