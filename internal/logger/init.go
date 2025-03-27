package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	// Configure encoder to remove stack traces and caller info
	config.EncoderConfig.StacktraceKey = "" // Remove stacktrace key
	config.EncoderConfig.CallerKey = ""     // Remove caller key
	config.EncoderConfig.NameKey = ""       // Remove name key
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"

	config.Level = zap.NewAtomicLevelAt(logLevel)

	zapLogger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	customLogger, err := NewCustomLogger(zapLogger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create custom logger: %v", err))
	}
	logger = customLogger
	logger.Info("Initializing logger",
		"environment", env,
		"log_level", logLevelStr,
		"debug", env == envDevelopment)
}
