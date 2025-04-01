// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envDevelopment = "development"
	defaultLevel   = "info"
)

// Module provides the logger module and its dependencies using fx.
var Module = fx.Options(
	fx.Provide(
		NewLogger,
	),
)

// NewLogger creates a new logger instance based on the provided configuration.
func NewLogger(cfg config.Interface) (types.Logger, error) {
	env := cfg.GetAppConfig().Environment
	if env == "" {
		env = envDevelopment
	}

	logConfig := cfg.GetLogConfig()
	logLevelStr := logConfig.Level
	if logLevelStr == "" {
		logLevelStr = defaultLevel
	}

	// If debug is enabled in app config, force debug level
	if cfg.GetAppConfig().Debug {
		logLevelStr = "debug"
	}

	logLevel, err := parseLogLevel(logLevelStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
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

	config := zap.NewProductionConfig()
	if env == envDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure encoder
	config.EncoderConfig.StacktraceKey = "" // Remove stacktrace key
	config.EncoderConfig.CallerKey = ""     // Remove caller key
	config.EncoderConfig.NameKey = ""       // Remove name key
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"

	// Create core based on environment
	var core zapcore.Core
	if env == envDevelopment {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(config.EncoderConfig),
			multiWriter,
			logLevel,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(config.EncoderConfig),
			multiWriter,
			logLevel,
		)
	}

	// Create logger
	zapLogger := zap.New(core)
	customLogger, err := NewCustomLogger(zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom logger: %w", err)
	}

	customLogger.Info("Initializing logger",
		"environment", env,
		"log_level", logLevelStr,
		"debug", env == envDevelopment)

	return customLogger, nil
}

// parseLogLevel converts a string log level to a zapcore.Level
func parseLogLevel(logLevelStr string) (zapcore.Level, error) {
	// Default to info level if no level is provided
	if logLevelStr == "" {
		return zapcore.InfoLevel, nil
	}

	switch logLevelStr {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.DebugLevel, fmt.Errorf("unknown log level: %s", logLevelStr)
	}
}
