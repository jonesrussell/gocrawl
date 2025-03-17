// Package logger provides logging functionality for the application.
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
	fx.Provide(provideLogger),
)

// InitializeLogger creates a new logger based on the configuration
func InitializeLogger(cfg config.Interface) (Interface, error) {
	logConfig := cfg.GetLogConfig()
	if logConfig.Debug {
		return NewDevelopmentLogger(logConfig.Level)
	}
	return NewProductionLogger(logConfig.Level)
}

// NewDevelopmentLogger initializes a new CustomLogger for development with colored output
func NewDevelopmentLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}

	// Development encoder config with colors for console
	devEncoderConfig := zap.NewDevelopmentEncoderConfig()
	devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Create file writer
	fileWriter, _, err := zap.Open("app.log")
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both console and file
	multiWriter := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(fileWriter),
	)

	// Console core with colored output
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(devEncoderConfig),
		multiWriter,
		logLevel,
	)

	// Create logger with both console and file output
	logger := zap.New(consoleCore, zap.AddCaller(), zap.Development())

	// Log when the logger is created
	logger.Info("Development logger initialized successfully")

	return NewCustomLogger(logger), nil
}

// NewProductionLogger initializes a new CustomLogger for production
func NewProductionLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}

	// Create file writer
	fileWriter, _, err := zap.Open("app.log")
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both console and file
	multiWriter := zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout),
		zapcore.AddSync(fileWriter),
	)

	// Use JSON encoder for both console and file logging
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		multiWriter,
		logLevel,
	)

	logger := zap.New(consoleCore)

	// Log when the logger is created
	logger.Info("Production logger initialized successfully")

	return NewCustomLogger(logger), nil
}

// parseLogLevel converts a string log level to a zapcore.Level
func parseLogLevel(logLevelStr string) (zapcore.Level, error) {
	var logLevel zapcore.Level

	// Default to info level if no level is provided
	if logLevelStr == "" {
		return zapcore.InfoLevel, nil
	}

	switch logLevelStr {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		return zapcore.DebugLevel, errors.New("unknown log level")
	}

	return logLevel, nil
}
