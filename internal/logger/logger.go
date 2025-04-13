// Package logger provides logging functionality for the application.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface defines the logger interface.
type Interface interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	With(fields ...any) Interface
}

// Logger implements the Interface.
type Logger struct {
	zapLogger *zap.Logger
}

var (
	// defaultLogger is the singleton logger instance
	defaultLogger *Logger
)

// New creates a new logger instance.
func New(config *Config) (Interface, error) {
	if defaultLogger != nil {
		return defaultLogger, nil
	}

	// Set default values
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Encoding == "" {
		config.Encoding = "console"
	}
	if len(config.OutputPaths) == 0 {
		config.OutputPaths = []string{"stdout"}
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Set output
	output := zapcore.AddSync(os.Stdout)

	// Create level enabler
	level := zapcore.InfoLevel
	if config.Development {
		level = zapcore.DebugLevel
	}

	// Create encoder based on encoding setting
	var encoder zapcore.Encoder
	if config.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		output,
		level,
	)

	// Create logger
	zapLogger := zap.New(core)
	defaultLogger = &Logger{zapLogger: zapLogger}

	return defaultLogger, nil
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a new logger with the given fields.
func (l *Logger) With(fields ...any) Interface {
	return &Logger{
		zapLogger: l.zapLogger.With(toZapFields(fields)...),
	}
}

// fieldPairSize represents the number of elements in a key-value pair.
const fieldPairSize = 2

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
