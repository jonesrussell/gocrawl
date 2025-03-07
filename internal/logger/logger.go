package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface holds the methods for the logger
type Interface interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Sync() error
}

// CustomLogger wraps the zap.Logger
type CustomLogger struct {
	*zap.Logger
	logFile *os.File
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

// InitializeLogger sets up the global logger based on the configuration
func InitializeLogger(cfg *config.Config) (Interface, error) {
	env := cfg.App.Environment
	if env == "" {
		env = "development" // Set a default environment
	}

	// Ensure we have a valid log level
	logLevel := cfg.Log.Level
	if logLevel == "" {
		logLevel = "info" // Set a default log level
	}

	// Log the initialization details
	fmt.Fprintf(os.Stderr, "Initializing logger in environment: %s with level: %s\n", env, logLevel)

	var customLogger *CustomLogger
	var err error
	if env == "development" {
		customLogger, err = NewDevelopmentLogger(logLevel)
	} else {
		customLogger, err = NewProductionLogger(logLevel)
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}

	return customLogger, nil
}

// Info logs an info message
func (c *CustomLogger) Info(msg string, fields ...interface{}) {
	c.Logger.Info(msg, convertToZapFields(fields)...)
}

// Error logs an error message
func (c *CustomLogger) Error(msg string, fields ...interface{}) {
	c.Logger.Error(msg, convertToZapFields(fields)...)
}

// Debug logs a debug message
func (c *CustomLogger) Debug(msg string, fields ...interface{}) {
	c.Logger.Debug(msg, convertToZapFields(fields)...)
}

// Warn logs a warning message
func (c *CustomLogger) Warn(msg string, fields ...interface{}) {
	c.Logger.Warn(msg, convertToZapFields(fields)...)
}

// Fatal logs a fatal message
func (c *CustomLogger) Fatal(msg string, fields ...interface{}) {
	c.Logger.Fatal(msg, convertToZapFields(fields)...)
}

// Printf logs a formatted message
func (c *CustomLogger) Printf(format string, args ...interface{}) {
	c.Logger.Info(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (c *CustomLogger) Errorf(format string, args ...interface{}) {
	c.Logger.Error(fmt.Sprintf(format, args...))
}

// Sync flushes any buffered log entries
func (c *CustomLogger) Sync() error {
	if err := c.Logger.Sync(); err != nil {
		return err
	}
	if c.logFile != nil {
		if err := c.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}
	return nil
}

// GetZapLogger returns the underlying zap.Logger
func (c *CustomLogger) GetZapLogger() *zap.Logger {
	return c.Logger
}

// convertToZapFields converts variadic key-value pairs to zap.Fields
func convertToZapFields(fields []interface{}) []zap.Field {
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
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("value%d", i), fields[i]))
			i-- // Adjust index since we're not consuming the next value
			continue
		}

		// Use the next item as value
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}

	// Handle last item if we have an odd number of fields
	if len(fields)%2 != 0 {
		last := fields[len(fields)-1]
		if str, ok := last.(string); ok {
			zapFields = append(zapFields, zap.String("context", str))
		} else {
			zapFields = append(zapFields, zap.Any("context", last))
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
