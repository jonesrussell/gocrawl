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
			var logger *CustomLogger
			var err error
			switch cfg.App.Environment {
			case "development":
				logger, err = NewDevelopmentLogger()
			case "production":
				logger, err = NewProductionLogger(cfg)
			default:
				err = errors.New("unknown environment")
			}
			if err != nil {
				panic(err)
			}
			return logger
		},
	),
)

// NewDevelopmentLogger initializes a new CustomLogger for development with colored output
func NewDevelopmentLogger() (*CustomLogger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color to log levels

	// Ensure the output is set to os.Stdout
	config.OutputPaths = []string{"stdout"} // This should work in most environments
	// Alternatively, you can use:
	// config.OutputPaths = []string{os.Stdout.Name()} // This is more explicit

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	// Log when the logger is created
	logger.Info("Development logger initialized successfully")
	return &CustomLogger{Logger: logger}, nil
}

// NewProductionLogger initializes a new CustomLogger for production
func NewProductionLogger(cfg *config.Config) (*CustomLogger, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

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
	core := zapcore.NewTee(fileCore, consoleCore)
	logger := zap.New(core)

	// Test logging to ensure it's working
	logger.Info("Production logger initialized successfully")

	return &CustomLogger{Logger: logger, logFile: logFile}, nil
}
