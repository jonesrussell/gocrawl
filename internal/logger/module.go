package logger

import (
	"errors"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger module and its dependencies
var Module = fx.Module("logger",
	fx.Provide(
		NewDevelopmentLogger, // Provide the development logger
		func() Interface { // Provide the logger.Interface
			logger, _ := NewDevelopmentLogger()
			return logger
		},
	),
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

// NewDevelopmentLogger initializes a new CustomLogger for development with colored output
func NewDevelopmentLogger() (*CustomLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color to log levels
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	// Log when the logger is created
	logger.Info("Logger initialized successfully")
	return &CustomLogger{Logger: logger}, nil
}

// NewProductionLogger initializes a new CustomLogger for production
func NewProductionLogger() (*CustomLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &CustomLogger{Logger: logger}, nil
}
