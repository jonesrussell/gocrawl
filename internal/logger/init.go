package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

const (
	envDevelopment = "development"
	defaultLevel   = "info"
)

var logger Interface

// GetLogger returns the global logger instance
func GetLogger() Interface {
	return logger
}

// Initialize logger with environment and log level
func init() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = envDevelopment
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = defaultLevel
	}

	logLevel, err := ParseLogLevel(logLevelStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse log level: %v", err))
	}

	config := zap.NewProductionConfig()
	if env == envDevelopment {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(logLevel)

	zapLogger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger = NewCustomLogger(zapLogger)
	logger.Info("Initializing logger",
		"environment", env,
		"log_level", logLevelStr,
		"debug", env == envDevelopment)
}
