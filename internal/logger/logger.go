package logger

import (
	"fmt"
	"os"

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
	Fatalf(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// CustomLogger wraps the zap.Logger
type CustomLogger struct {
	Logger *zap.Logger
	Level  zapcore.Level
}

// Ensure CustomLogger implements Interface
var _ Interface = (*CustomLogger)(nil)

// Params holds the parameters for creating a logger
type Params struct {
	fx.In

	Level  zapcore.Level
	AppEnv string `name:"appEnv"`
}

// NewCustomLogger initializes a new CustomLogger with a specified log level
func NewCustomLogger(params Params) (*CustomLogger, error) {
	config := zap.Config{
		Level:    zap.NewAtomicLevelAt(params.Level),
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			LevelKey:      "level",
			TimeKey:       "time",
			CallerKey:     "caller",
			StacktraceKey: "stacktrace",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &CustomLogger{Logger: logger, Level: params.Level}, nil
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

// Fatalf logs a fatal message
func (c *CustomLogger) Fatalf(msg string, args ...interface{}) {
	c.Logger.Fatal(msg, zap.Any("args", args))
}

// Errorf logs a formatted error message
func (c *CustomLogger) Errorf(format string, args ...interface{}) {
	c.Logger.Error(fmt.Sprintf(format, args...))
}

// NewDevelopmentLogger initializes a new CustomLogger for development
func NewDevelopmentLogger(p Params) (*CustomLogger, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel,
	)

	logger := zap.New(core)
	return &CustomLogger{Logger: logger, Level: p.Level}, nil
}

// Sync flushes any buffered log entries
func (c *CustomLogger) Sync() error {
	return c.Logger.Sync()
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

	// If first argument is a string and no more arguments, treat it as additional message context
	if len(fields) == 1 {
		if str, ok := fields[0].(string); ok {
			return []zap.Field{zap.String("context", str)}
		}
	}

	// Handle key-value pairs
	for i := 0; i < len(fields); i++ {
		// If we have an odd number of remaining fields, add the last one as context
		if i == len(fields)-1 {
			if str, ok := fields[i].(string); ok {
				zapFields = append(zapFields, zap.String("context", str))
			} else {
				zapFields = append(zapFields, zap.Any("context", fields[i]))
			}
			break
		}

		// Process key-value pair
		key, ok := fields[i].(string)
		if !ok {
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("value%d", i), fields[i]))
			continue
		}

		// Move to next item for value
		i++
		value := fields[i]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
