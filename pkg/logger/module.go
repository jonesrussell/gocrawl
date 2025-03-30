// Package logger provides logging functionality for the application.
package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger module for dependency injection.
var Module = fx.Module("logger",
	fx.Provide(
		// Provide the logger with configuration
		fx.Annotate(
			NewLogger,
			fx.As(new(Interface)),
		),
	),
)

// NewLogger creates a new logger instance with the given configuration.
func NewLogger(p Params) (*zap.Logger, error) {
	// Create the logger configuration
	config := zap.NewProductionConfig()
	if p.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Create the logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
