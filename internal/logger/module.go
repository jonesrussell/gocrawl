package logger

import (
	"errors"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger module and its dependencies
var Module = fx.Module("logger",
	fx.Provide(func(cfg *config.Config) Interface {
		logger, err := InitializeLogger(cfg)
		if err != nil {
			panic(err)
		}
		return logger
	}),
)

// NewDevelopmentLogger initializes a new CustomLogger for development with colored output
func NewDevelopmentLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}

	// Development encoder config with colors for console
	devEncoderConfig := zap.NewDevelopmentEncoderConfig()
	devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Console core with colored output
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(devEncoderConfig),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// Create logger without file output for testing
	logger := zap.New(consoleCore, zap.AddCaller(), zap.Development())

	// Log when the logger is created
	logger.Info("Development logger initialized successfully")

	return &CustomLogger{Logger: logger}, nil
}

// NewProductionLogger initializes a new CustomLogger for production
func NewProductionLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, err
	}

	// Use JSON encoder for console logging
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	logger := zap.New(consoleCore)

	// Log when the logger is created
	logger.Info("Production logger initialized successfully")

	return &CustomLogger{Logger: logger}, nil
}

// parseLogLevel converts a string log level to a zapcore.Level
func parseLogLevel(logLevelStr string) (zapcore.Level, error) {
	var logLevel zapcore.Level

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
