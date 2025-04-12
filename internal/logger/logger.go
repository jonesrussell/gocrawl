// Package logger provides logging functionality for the application.
package logger

import (
	"os"

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
	// Set default values if not provided
	if config == nil {
		config = DefaultConfig()
	}

	// Set the log level
	zapLevel := levelToZap(config.Level)

	// Set the encoding
	if config.Encoding == "" {
		config.Encoding = "console"
	}

	// Set the output paths
	if len(config.OutputPaths) == 0 {
		config.OutputPaths = []string{"stdout"}
	}

	// Set the error output paths
	if len(config.ErrorOutputPaths) == 0 {
		config.ErrorOutputPaths = []string{"stderr"}
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
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configure level encoder based on development mode and color settings
	if config.EnableColor {
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

	// Create core with debug level if in development mode
	var core zapcore.Core
	if config.Encoding == "json" {
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
	if config.Development {
		zapLogger = zap.New(core, zap.Development(), zap.AddCaller())
	} else {
		zapLogger = zap.New(core, zap.AddCaller())
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
