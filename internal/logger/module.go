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
		func(cfg *config.Config) Interface { // Provide the logger.Interface
			var customLogger *CustomLogger
			var err error

			switch cfg.App.Environment {
			case "development":
				customLogger, err = NewDevelopmentLogger(cfg.Log.Level) // Pass log level as string
			case "production":
				customLogger, err = NewProductionLogger(cfg.Log.Level) // Pass log level as string
			default:
				err = errors.New("unknown environment")
			}
			if err != nil {
				panic(err)
			}
			return customLogger
		},
	),
)

// NewDevelopmentLogger initializes a new CustomLogger for development with colored output
func NewDevelopmentLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel := zapcore.InfoLevel // Default log level

	// Convert string log level to zapcore.Level
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
		return nil, errors.New("unknown log level")
	}

	devCfg := zap.NewDevelopmentConfig()
	devCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color to log levels
	devCfg.Level = zap.NewAtomicLevelAt(logLevel)                       // Set log level

	// Ensure the output is set to os.Stdout
	devCfg.OutputPaths = []string{"stdout"} // This should work in most environments

	log, err := devCfg.Build()
	if err != nil {
		return nil, err
	}
	// Log when the logger is created
	log.Info("Development logger initialized successfully")
	return &CustomLogger{Logger: log}, nil
}

// NewProductionLogger initializes a new CustomLogger for production
func NewProductionLogger(logLevelStr string) (*CustomLogger, error) {
	logLevel := zapcore.InfoLevel // Default log level

	// Convert string log level to zapcore.Level
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
		return nil, errors.New("unknown log level")
	}

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Use JSON encoder for file logging
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(logFile),
		zapcore.Level(logLevel), // Set log level for file logging
	)

	// Use JSON encoder for console logging as well
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), // Change to JSON
		zapcore.AddSync(os.Stdout),
		logLevel, // Set console log level to match the desired log level
	)

	core := zapcore.NewTee(fileCore, consoleCore)
	logger := zap.New(core)

	// Log when the logger is created
	logger.Info("Production logger initialized successfully")

	return &CustomLogger{Logger: logger, logFile: logFile}, nil
}
