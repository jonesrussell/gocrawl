// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	// Structured logging helpers
	WithUser(userID string) Interface
	WithRequestID(requestID string) Interface
	WithTraceID(traceID string) Interface
	WithSpanID(spanID string) Interface
	WithDuration(duration time.Duration) Interface
	WithError(err error) Interface
	WithComponent(component string) Interface
	WithVersion(version string) Interface
	WithEnvironment(env string) Interface
}

// Logger implements the Interface.
type Logger struct {
	zapLogger *zap.Logger
}

var (
	// defaultLogger is the singleton logger instance
	defaultLogger *Logger

	// logLevels maps string levels to zapcore.Level
	logLevels = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
		"fatal": zapcore.FatalLevel,
	}

	// Common field keys
	fieldKeys = struct {
		UserID      string
		RequestID   string
		TraceID     string
		SpanID      string
		Duration    string
		Error       string
		Component   string
		Version     string
		Environment string
	}{
		UserID:      "user_id",
		RequestID:   "request_id",
		TraceID:     "trace_id",
		SpanID:      "span_id",
		Duration:    "duration",
		Error:       "error",
		Component:   "component",
		Version:     "version",
		Environment: "environment",
	}
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
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	if config.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		}
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoderConfig.ConsoleSeparator = " | "
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Create encoder
	var encoder zapcore.Encoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Get log level
	level := getLogLevel(string(config.Level))

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger with options
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if config.Development {
		opts = append(opts, zap.Development())
	}
	zapLogger := zap.New(core, opts...)

	defaultLogger = &Logger{zapLogger: zapLogger}
	return defaultLogger, nil
}

// getLogLevel converts a string level to zapcore.Level
func getLogLevel(level string) zapcore.Level {
	lvl, exists := logLevels[strings.ToLower(level)]
	if !exists {
		return zapcore.InfoLevel
	}
	return lvl
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

// WithUser adds a user ID to the logger.
func (l *Logger) WithUser(userID string) Interface {
	return l.With(fieldKeys.UserID, userID)
}

// WithRequestID adds a request ID to the logger.
func (l *Logger) WithRequestID(requestID string) Interface {
	return l.With(fieldKeys.RequestID, requestID)
}

// WithTraceID adds a trace ID to the logger.
func (l *Logger) WithTraceID(traceID string) Interface {
	return l.With(fieldKeys.TraceID, traceID)
}

// WithSpanID adds a span ID to the logger.
func (l *Logger) WithSpanID(spanID string) Interface {
	return l.With(fieldKeys.SpanID, spanID)
}

// WithDuration adds a duration to the logger.
func (l *Logger) WithDuration(duration time.Duration) Interface {
	return l.With(fieldKeys.Duration, duration)
}

// WithError adds an error to the logger.
func (l *Logger) WithError(err error) Interface {
	return l.With(fieldKeys.Error, err)
}

// WithComponent adds a component name to the logger.
func (l *Logger) WithComponent(component string) Interface {
	return l.With(fieldKeys.Component, component)
}

// WithVersion adds a version to the logger.
func (l *Logger) WithVersion(version string) Interface {
	return l.With(fieldKeys.Version, version)
}

// WithEnvironment adds an environment to the logger.
func (l *Logger) WithEnvironment(env string) Interface {
	return l.With(fieldKeys.Environment, env)
}

// toZapFields converts a list of any fields to zap.Field.
func toZapFields(fields []any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		switch field := fields[i].(type) {
		case zap.Field:
			// If it's already a zap.Field, use it directly
			zapFields = append(zapFields, field)
		case string:
			// If it's a string, it should be a key
			if i+1 >= len(fields) {
				if defaultLogger != nil {
					defaultLogger.Warn("Missing value for field key",
						"key", field,
						"error", ErrInvalidFields,
					)
				}
				continue
			}
			zapFields = append(zapFields, zap.Any(field, fields[i+1]))
			i++ // Skip the value in the next iteration
		default:
			// If it's neither, log a warning and skip
			if defaultLogger != nil {
				defaultLogger.Warn("Invalid field type",
					"expected_type", "string or zap.Field",
					"actual_type", fmt.Sprintf("%T", field),
					"error", ErrInvalidFields,
				)
			}
		}
	}

	return zapFields
}
