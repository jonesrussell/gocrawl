package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// CustomLogger wraps the zap.Logger
type CustomLogger struct {
	*zap.Logger
	fatalHook func(zapcore.Entry) error
}

// Ensure CustomLogger implements Interface
var _ Interface = (*CustomLogger)(nil)

// Params holds the parameters for creating a logger
type Params struct {
	fx.In

	Debug  bool
	Level  zapcore.Level
	AppEnv string `name:"appEnv"`
}

const (
	defaultLogLevel = "info"
)

// Info logs an info message
func (c *CustomLogger) Info(msg string, fields ...any) {
	c.Logger.Info(msg, ConvertToZapFields(fields)...)
}

// Error logs an error message
func (c *CustomLogger) Error(msg string, fields ...any) {
	c.Logger.Error(msg, ConvertToZapFields(fields)...)
}

// Debug logs a debug message
func (c *CustomLogger) Debug(msg string, fields ...any) {
	c.Logger.Debug(msg, ConvertToZapFields(fields)...)
}

// Warn logs a warning message
func (c *CustomLogger) Warn(msg string, fields ...any) {
	c.Logger.Warn(msg, ConvertToZapFields(fields)...)
}

// Fatal logs a fatal message and executes the fatal hook
func (c *CustomLogger) Fatal(msg string, fields ...any) {
	if c.fatalHook != nil {
		entry := zapcore.Entry{
			Level:   zapcore.FatalLevel,
			Message: msg,
		}
		_ = c.fatalHook(entry)
	}
	c.Logger.Fatal(msg, ConvertToZapFields(fields)...)
}

// Printf logs a formatted message
func (c *CustomLogger) Printf(format string, args ...any) {
	c.Logger.Info(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (c *CustomLogger) Errorf(format string, args ...any) {
	c.Logger.Error(fmt.Sprintf(format, args...))
}

// Sync flushes any buffered log entries
func (c *CustomLogger) Sync() error {
	return c.Logger.Sync()
}

// GetZapLogger returns the underlying zap.Logger
func (c *CustomLogger) GetZapLogger() *zap.Logger {
	return c.Logger
}

// ParseLogLevel converts a string log level to a zapcore.Level
func ParseLogLevel(logLevelStr string) (zapcore.Level, error) {
	var logLevel zapcore.Level

	// If no level specified, use default
	if logLevelStr == "" {
		logLevelStr = defaultLogLevel
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
		return zapcore.DebugLevel, fmt.Errorf("unknown log level: %s", logLevelStr)
	}

	return logLevel, nil
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

// NewCustomLogger creates a new CustomLogger. If a logger is provided, it will be used.
// Otherwise, a new logger will be created with default configuration.
func NewCustomLogger(logger *zap.Logger) (*CustomLogger, error) {
	if logger != nil {
		return &CustomLogger{
			Logger: logger,
			fatalHook: func(entry zapcore.Entry) error {
				os.Exit(1)
				return nil
			},
		}, nil
	}

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "" // Remove stacktrace key
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "" // Remove caller key
	config.EncoderConfig.NameKey = ""   // Remove name key

	// Disable stacktrace for warnings
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	newLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &CustomLogger{
		Logger: newLogger,
		fatalHook: func(entry zapcore.Entry) error {
			os.Exit(1)
			return nil
		},
	}, nil
}

// SetFatalHook allows overriding the default fatal behavior for testing
func (c *CustomLogger) SetFatalHook(hook func(zapcore.Entry) error) {
	c.fatalHook = hook
}

// NewNoOp creates a no-op logger that discards all logs
func NewNoOp() Interface {
	return &noOpLogger{}
}

// noOpLogger implements Interface but discards all logs
type noOpLogger struct{}

func (l *noOpLogger) Debug(string, ...any)  {}
func (l *noOpLogger) Info(string, ...any)   {}
func (l *noOpLogger) Warn(string, ...any)   {}
func (l *noOpLogger) Error(string, ...any)  {}
func (l *noOpLogger) Fatal(string, ...any)  {}
func (l *noOpLogger) Printf(string, ...any) {}
func (l *noOpLogger) Errorf(string, ...any) {}
func (l *noOpLogger) Sync() error           { return nil }
