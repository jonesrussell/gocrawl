// Package logger provides logging functionality for the application.
package logger

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Printf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Sync() error
}

// Config holds logger configuration.
type Config struct {
	Level  string
	Debug  bool
	Output string
}

// DefaultConfig returns default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:  "info",
		Debug:  false,
		Output: "stdout",
	}
}

// zapLogger implements Logger using zap.
type zapLogger struct {
	*zap.Logger
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.Logger.Info(msg, convertFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.Logger.Error(msg, convertFields(fields)...)
}

func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.Logger.Debug(msg, convertFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.Logger.Warn(msg, convertFields(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.Logger.Fatal(msg, convertFields(fields)...)
}

func (l *zapLogger) Printf(format string, v ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, v...))
}

func (l *zapLogger) Errorf(format string, v ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, v...))
}

// New creates a new logger with the given configuration.
func New(cfg Config) (Logger, error) {
	config := zap.NewProductionConfig()
	if cfg.Debug {
		config = zap.NewDevelopmentConfig()
	}

	// Set output
	if cfg.Output != "stdout" {
		config.OutputPaths = []string{cfg.Output}
	}

	// Set level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Configure encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	config.EncoderConfig.ConsoleSeparator = " | "

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &zapLogger{Logger: logger}, nil
}

// NewNoOp creates a no-op logger that discards all messages.
func NewNoOp() Logger {
	return &noOpLogger{}
}

// noOpLogger implements Logger but discards all messages.
type noOpLogger struct{}

func (l *noOpLogger) Info(msg string, fields ...interface{})  {}
func (l *noOpLogger) Error(msg string, fields ...interface{}) {}
func (l *noOpLogger) Debug(msg string, fields ...interface{}) {}
func (l *noOpLogger) Warn(msg string, fields ...interface{})  {}
func (l *noOpLogger) Fatal(msg string, fields ...interface{}) {}
func (l *noOpLogger) Printf(format string, v ...interface{})  {}
func (l *noOpLogger) Errorf(format string, v ...interface{})  {}
func (l *noOpLogger) Sync() error                             { return nil }

// convertFields converts variadic key-value pairs to zap fields.
func convertFields(fields []interface{}) []zap.Field {
	var zapFields []zap.Field

	if len(fields) == 0 {
		return zapFields
	}

	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("value%d", i), maskSensitiveData(fields[i])))
			i--
			continue
		}
		zapFields = append(zapFields, zap.Any(key, maskSensitiveData(fields[i+1])))
	}

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

// maskSensitiveData masks sensitive information in the given value.
func maskSensitiveData(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		masked := make(map[string]interface{})
		for key, val := range v {
			if isSensitiveField(key) {
				masked[key] = "[REDACTED]"
			} else {
				masked[key] = maskSensitiveData(val)
			}
		}
		return masked
	case []interface{}:
		masked := make([]interface{}, len(v))
		for i, val := range v {
			masked[i] = maskSensitiveData(val)
		}
		return masked
	default:
		return value
	}
}

// isSensitiveField checks if a field name indicates sensitive data.
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

// WithContext adds a logger to the context.
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext retrieves the logger from the context.
func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(contextKey{}).(Logger)
	if !ok {
		return NewNoOp()
	}
	return logger
}

type contextKey struct{}

// logger implements the Interface using zap.
type logger struct {
	*zap.Logger
	config *Config
}

// Debug logs a debug message.
func (l *logger) Debug(msg string, fields ...interface{}) {
	l.Logger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message.
func (l *logger) Info(msg string, fields ...interface{}) {
	l.Logger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *logger) Warn(msg string, fields ...interface{}) {
	l.Logger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message.
func (l *logger) Error(msg string, fields ...interface{}) {
	l.Logger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *logger) Fatal(msg string, fields ...interface{}) {
	l.Logger.Fatal(msg, toZapFields(fields)...)
}

// With creates a child logger with additional fields.
func (l *logger) With(fields ...interface{}) Interface {
	return &logger{
		Logger: l.Logger.With(toZapFields(fields)...),
		config: l.config,
	}
}

// toZapFields converts a list of interface{} fields to zap.Field.
func toZapFields(fields []interface{}) []zap.Field {
	if len(fields)%2 != 0 {
		return []zap.Field{zap.Error(ErrInvalidFields)}
	}

	zapFields := make([]zap.Field, 0, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(ErrInvalidFields)}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
