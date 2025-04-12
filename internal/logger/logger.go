// Package logger provides logging functionality for the application.
package logger

import (
	"os"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
func New(cfg *Config) (Interface, error) {
	// Set default values if not provided
	if cfg == nil {
		cfg = DefaultConfig()
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
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configure level encoder based on development mode and color settings
	if cfg.EnableColor {
		encoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			var color string
			switch l {
			case zapcore.DebugLevel:
				color = ColorDebug
			case zapcore.InfoLevel:
				color = ColorInfo
			case zapcore.WarnLevel:
				color = ColorWarn
			case zapcore.ErrorLevel:
				color = ColorError
			case zapcore.DPanicLevel:
				color = ColorError
			case zapcore.PanicLevel:
				color = ColorError
			case zapcore.FatalLevel:
				color = ColorFatal
			case zapcore.InvalidLevel:
				color = ColorReset
			default:
				color = ColorReset
			}
			enc.AppendString(color + l.CapitalString() + ColorReset)
		}
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// Convert logger level to zapcore level
	zapLevel := levelToZap(cfg.Level)

	// Create core
	var core zapcore.Core
	if cfg.Encoding == "json" {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	}

	// Create logger with development mode if enabled
	var zapLogger *zap.Logger
	if cfg.Development {
		zapLogger = zap.New(core, zap.Development(), zap.AddCaller())
	} else {
		zapLogger = zap.New(core, zap.AddCaller())
	}

	return &logger{
		zapLogger: zapLogger,
		config:    cfg,
	}, nil
}

// fieldPairSize represents the number of elements in a key-value pair.
const fieldPairSize = 2

// logger implements the Interface.
type logger struct {
	zapLogger *zap.Logger
	config    *Config
}

// Debug logs a debug message.
func (l *logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message.
func (l *logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message.
func (l *logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a child logger with additional fields.
func (l *logger) With(fields ...any) Interface {
	return &logger{
		zapLogger: l.zapLogger.With(toZapFields(fields)...),
		config:    l.config,
	}
}

// NewFxLogger creates a new Fx logger.
func (l *logger) NewFxLogger() fxevent.Logger {
	return NewFxLogger(l.zapLogger)
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
