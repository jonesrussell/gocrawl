// Package logger provides logging functionality for the application.
package logger

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
	levelFatal = "fatal"
)

// Interface defines the interface for logging operations.
// It provides structured logging capabilities with different log levels and
// support for additional fields in log messages.
type Interface interface {
	// Debug logs a debug message with optional fields.
	// Used for detailed information useful during development.
	Debug(msg string, fields ...any)
	// Error logs an error message with optional fields.
	// Used for error conditions that need immediate attention.
	Error(msg string, fields ...any)
	// Info logs an informational message with optional fields.
	// Used for general operational information.
	Info(msg string, fields ...any)
	// Warn logs a warning message with optional fields.
	// Used for potentially harmful situations.
	Warn(msg string, fields ...any)
	// Fatal logs a fatal message and panics.
	// Used for unrecoverable errors that require immediate termination.
	Fatal(msg string, fields ...any)
	// Printf logs a formatted message.
	// Used for formatted string logging.
	Printf(format string, args ...any)
	// Errorf logs a formatted error message.
	// Used for formatted error string logging.
	Errorf(format string, args ...any)
	// Sync flushes any buffered log entries.
	// Used to ensure all logs are written before shutdown.
	Sync() error
}

// Params holds the parameters for creating a logger
type Params struct {
	Debug  bool
	Level  string
	AppEnv string
}

// ZapLogger implements Interface using zap.Logger
type ZapLogger struct {
	*zap.Logger
}

func (l *ZapLogger) Debug(msg string, fields ...any) {
	l.Logger.Debug(msg, ConvertToZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...any) {
	l.Logger.Error(msg, ConvertToZapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...any) {
	l.Logger.Info(msg, ConvertToZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...any) {
	l.Logger.Warn(msg, ConvertToZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...any) {
	l.Logger.Fatal(msg, ConvertToZapFields(fields)...)
}

func (l *ZapLogger) Printf(format string, args ...any) {
	l.Logger.Sugar().Infof(format, args...)
}

func (l *ZapLogger) Errorf(format string, args ...any) {
	l.Logger.Sugar().Errorf(format, args...)
}

func (l *ZapLogger) Sync() error {
	return l.Logger.Sync()
}

// NewNoOp creates a no-op logger that discards all log messages.
// This is useful for testing or when logging is not needed.
func NewNoOp() Interface {
	return &NoOpLogger{}
}

// NoOpLogger implements Interface but discards all log messages.
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, fields ...any)   {}
func (l *NoOpLogger) Error(msg string, fields ...any)   {}
func (l *NoOpLogger) Info(msg string, fields ...any)    {}
func (l *NoOpLogger) Warn(msg string, fields ...any)    {}
func (l *NoOpLogger) Fatal(msg string, fields ...any)   {}
func (l *NoOpLogger) Printf(format string, args ...any) {}
func (l *NoOpLogger) Errorf(format string, args ...any) {}
func (l *NoOpLogger) Sync() error                       { return nil }

// NewTestLogger creates a new logger for testing.
func NewTestLogger() Interface {
	logger, _ := zap.NewDevelopment()
	return &ZapLogger{Logger: logger}
}

// maskSensitiveData masks sensitive information in the given value
func maskSensitiveData(value any) any {
	switch v := value.(type) {
	case map[string]any:
		masked := make(map[string]any)
		for key, val := range v {
			// Mask sensitive fields
			if isSensitiveField(key) {
				masked[key] = "[REDACTED]"
			} else {
				masked[key] = maskSensitiveData(val)
			}
		}
		return masked
	case []any:
		masked := make([]any, len(v))
		for i, val := range v {
			masked[i] = maskSensitiveData(val)
		}
		return masked
	default:
		return value
	}
}

// isSensitiveField checks if a field name indicates sensitive data
func isSensitiveField(field string) bool {
	sensitiveFields := []string{
		"password",
		"apiKey",
		"apikey",
		"token",
		"secret",
		"key",
		"credentials",
	}
	for _, s := range sensitiveFields {
		if strings.Contains(strings.ToLower(field), s) {
			return true
		}
	}
	return false
}

// ConvertToZapFields converts variadic key-value pairs to zap.Fields
func ConvertToZapFields(fields []any) []zap.Field {
	var zapFields []zap.Field

	// If no fields provided, return empty slice
	if len(fields) == 0 {
		return zapFields
	}

	// Handle key-value pairs
	for i := 0; i < len(fields)-1; i += 2 {
		// Process key-value pair
		key, ok := fields[i].(string)
		if !ok {
			// If key is not a string, use it as a value with a generated key
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("value%d", i), maskSensitiveData(fields[i])))
			i-- // Adjust index since we're not consuming the next value
			continue
		}

		// Use the next item as value
		zapFields = append(zapFields, zap.Any(key, maskSensitiveData(fields[i+1])))
	}

	// Handle last item if we have an odd number of fields
	if len(fields)%2 != 0 {
		last := fields[len(fields)-1]
		if str, ok := last.(string); ok {
			zapFields = append(zapFields, zap.String("context", str))
		} else {
			zapFields = append(zapFields, zap.Any("context", maskSensitiveData(last)))
		}
	}

	return zapFields
}

type contextKey struct{}

// WithContext adds a logger to the context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext retrieves the logger from the context
func FromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(contextKey{}).(*zap.Logger)
	if !ok {
		// Return a default logger or handle the error as needed
		return zap.NewNop() // No-op logger
	}
	return logger
}

// NewCustomLogger creates a new logger with the given parameters.
// If a logger is provided, it will be used. Otherwise, a new logger will be created
// with the given configuration.
func NewCustomLogger(logger *zap.Logger, params Params) (Interface, error) {
	if logger != nil {
		return &ZapLogger{
			Logger: logger,
		}, nil
	}

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Set log level based on params
	var level zapcore.Level
	switch params.Level {
	case levelDebug:
		level = zapcore.DebugLevel
	case levelInfo:
		level = zapcore.InfoLevel
	case levelWarn:
		level = zapcore.WarnLevel
	case levelError:
		level = zapcore.ErrorLevel
	case levelFatal:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Create the logger
	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{
		Logger: zapLogger,
	}, nil
}
