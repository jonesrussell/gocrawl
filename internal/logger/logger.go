// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface defines the logging interface.
type Interface interface {
	// Debug logs a debug message with optional fields.
	Debug(msg string, fields ...any)
	// Info logs an info message with optional fields.
	Info(msg string, fields ...any)
	// Warn logs a warning message with optional fields.
	Warn(msg string, fields ...any)
	// Error logs an error message with optional fields.
	Error(msg string, fields ...any)
	// With creates a new logger with the given fields.
	With(fields ...any) Interface
}

// DefaultConfig returns a default configuration for the logger.
func DefaultConfig() *Config {
	return &Config{
		Level:            InfoLevel,
		Development:      true,
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EnableColor:      true,
	}
}

// New creates a new logger with the given configuration.
func New(config *Config) (Interface, error) {
	zapConfig := zap.NewProductionConfig()
	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set the log level
	zapLevel := levelToZap(config.Level)
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	// Set the encoding
	zapConfig.Encoding = config.Encoding

	// Set the output paths
	zapConfig.OutputPaths = config.OutputPaths

	// Set the error output paths
	zapConfig.ErrorOutputPaths = config.ErrorOutputPaths

	// Set the encoder config
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Build the logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// Wrap the zap logger in our custom logger type
	return &logger{zapLogger: zapLogger}, nil
}

// fieldPairSize represents the number of elements in a key-value pair.
const fieldPairSize = 2

// logger implements the Interface using zap.Logger.
type logger struct {
	zapLogger *zap.Logger
}

// Debug logs a debug message with optional fields.
func (l *logger) Debug(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	l.zapLogger.Debug(msg, zapFields...)
}

// Info logs an info message with optional fields.
func (l *logger) Info(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning message with optional fields.
func (l *logger) Warn(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error message with optional fields.
func (l *logger) Error(msg string, fields ...any) {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	l.zapLogger.Error(msg, zapFields...)
}

// Fatal logs a fatal message and exits.
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a new logger with the given fields.
func (l *logger) With(fields ...any) Interface {
	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return &logger{
		zapLogger: l.zapLogger.With(zapFields...),
	}
}

// toZapFields converts a list of any fields to zap.Field.
func toZapFields(fields []any) []zap.Field {
	if len(fields)%fieldPairSize != 0 {
		return []zap.Field{zap.Error(ErrInvalidFields)}
	}

	zapFields := make([]zap.Field, 0, len(fields)/fieldPairSize)
	for i := 0; i < len(fields); i += fieldPairSize {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(ErrInvalidFields)}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
