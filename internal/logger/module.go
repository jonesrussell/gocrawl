// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Params contains the dependencies required to create a logger.
type Params struct {
	fx.In

	Config *Config `optional:"true"`
}

// Result contains the logger instance.
type Result struct {
	fx.Out

	Logger Interface
}

// Module provides the logger module's dependencies.
var Module = fx.Module("logger",
	fx.Provide(
		// Provide the logger instance
		func(p Params) (Result, error) {
			// Create default config if not provided
			config := p.Config
			if config == nil {
				config = &Config{
					Level:            DefaultLevel,
					Development:      DefaultDevelopment,
					Encoding:         DefaultEncoding,
					OutputPaths:      DefaultOutputPaths,
					ErrorOutputPaths: DefaultErrorOutputPaths,
				}
			}

			// Create zap logger
			zapConfig := zap.NewProductionConfig()
			if config.Development {
				zapConfig = zap.NewDevelopmentConfig()
			}

			zapConfig.Level = zap.NewAtomicLevelAt(levelToZap(config.Level))
			zapConfig.Encoding = config.Encoding
			zapConfig.OutputPaths = config.OutputPaths
			zapConfig.ErrorOutputPaths = config.ErrorOutputPaths

			zapLogger, err := zapConfig.Build()
			if err != nil {
				return Result{}, fmt.Errorf("failed to create logger: %w", err)
			}

			// Create logger instance
			logger := &logger{
				Logger: zapLogger,
				config: config,
			}

			return Result{Logger: logger}, nil
		},
	),
)

// levelToZap converts a logger.Level to a zapcore.Level.
func levelToZap(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
