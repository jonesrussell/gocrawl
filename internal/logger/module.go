// Package logger provides logging functionality for the application.
package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Params holds the parameters for creating a new logger.
type Params struct {
	Config *Config
}

// Result contains the logger instance.
type Result struct {
	fx.Out

	Logger Interface
}

// Module provides the logger module.
var Module = fx.Module("logger",
	fx.Provide(
		fx.Annotate(
			Constructor,
			fx.As(new(Interface)),
		),
	),
)

// Constructor creates a new logger instance.
func Constructor(params Params) (Interface, error) {
	config := params.Config
	if config == nil {
		config = &Config{
			Level:       InfoLevel,
			Development: false,
			Encoding:    "json",
		}
	}

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
		return nil, err
	}

	return &logger{
		zapLogger: zapLogger,
		config:    config,
	}, nil
}

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
