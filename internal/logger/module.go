package logger

import (
	"errors"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger module and its dependencies
var Module = fx.Module("logger",
	fx.Provide(
		NewDevelopmentLogger, // Provide the development logger
		func(cfg *config.Config) Interface { // Provide the logger.Interface
			logger, _ := NewLogger(cfg) // Use the new logger function
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
	var logFile *os.File

	// Create a log file for development
	if cfg.App.Environment == "development" {
		logFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	var core zapcore.Core
	switch cfg.App.Environment {
	case "development":
		// Development logger with file output
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(logFile),
			zapcore.DebugLevel, // Log level for file
		)
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel, // Log level for console
		)
		core = zapcore.NewTee(fileCore, consoleCore)
		logger = zap.New(core)
	case "staging":
		// Staging logger
		config := zap.NewDevelopmentConfig()
		logger, err = config.Build()
	case "production":
		// Production logger with file output
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(logFile),
			zapcore.InfoLevel,
		)
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		core = zapcore.NewTee(fileCore, consoleCore)
		logger = zap.New(core)
	default:
		return nil, errors.New("unknown environment")
	}

	if err != nil {
		return nil, err
	}

	// Test logging to ensure it's working
	logger.Info("Logger initialized successfully")

	return &CustomLogger{logger, logFile}, nil
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
