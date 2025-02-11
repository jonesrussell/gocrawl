package logger

import (
	"errors"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger as an Fx module
//
//nolint:gochecknoglobals // This is a module definition
var Module = fx.Module("logger",
	fx.Provide(NewLogger),
)

// NewLogger initializes the appropriate logger based on the environment
func NewLogger(cfg *config.Config) (*CustomLogger, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	var logger *zap.Logger
	var err error
	switch cfg.App.Environment {
	case "development":
		// Development logger with color
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color to log levels
		logger, err = config.Build()
	case "staging":
		// Staging logger
		config := zap.NewDevelopmentConfig()
		logger, err = config.Build()
	case "production":
		// Production logger
		config := zap.NewProductionConfig()
		logger, err = config.Build()
	default:
		return nil, errors.New("unknown environment")
	}

	if err != nil {
		return nil, err
	}

	return &CustomLogger{logger}, nil
}
