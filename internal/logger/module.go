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

// Module provides the logger module and its dependencies using fx.
var Module = fx.Options(
	fx.Provide(
		provideLogger,
	),
)

// Impl implements the Interface
type Impl struct {
	logger *zap.Logger
}

// Ensure Impl implements Interface
var _ Interface = (*Impl)(nil)

// NewLogger creates a new logger instance
func NewLogger(cfg *config.LogConfig) (Interface, error) {
	logLevel, err := parseLogLevel(cfg.Level)
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

	// Create core based on environment
	var core zapcore.Core
	if cfg.Debug {
		devEncoderConfig := zap.NewDevelopmentEncoderConfig()
		devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(devEncoderConfig),
			multiWriter,
			logLevel,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			multiWriter,
			logLevel,
		)
	}

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.Development())
	return &Impl{
		logger: zapLogger,
	}, nil
}

// Info implements Interface
func (l *Impl) Info(msg string, fields ...any) {
	l.logger.Info(msg, ConvertToZapFields(fields)...)
}

// Error implements Interface
func (l *Impl) Error(msg string, fields ...any) {
	l.logger.Error(msg, ConvertToZapFields(fields)...)
}

// Debug implements Interface
func (l *Impl) Debug(msg string, fields ...any) {
	l.logger.Debug(msg, ConvertToZapFields(fields)...)
}

// Warn implements Interface
func (l *Impl) Warn(msg string, fields ...any) {
	l.logger.Warn(msg, ConvertToZapFields(fields)...)
}

// Fatal implements Interface
func (l *Impl) Fatal(msg string, fields ...any) {
	l.logger.Fatal(msg, ConvertToZapFields(fields)...)
}

// Printf implements Interface
func (l *Impl) Printf(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Errorf implements Interface
func (l *Impl) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

// Sync implements Interface
func (l *Impl) Sync() error {
	return l.logger.Sync()
}

// provideLogger creates a new logger instance
func provideLogger() (Interface, error) {
	return GetLogger(), nil
}

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
