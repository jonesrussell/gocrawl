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
	// Create a zap configuration
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

	// Create the logger
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

// Implement Fatalf method
func (c *CustomLogger) Fatalf(msg string, args ...interface{}) {
	c.Logger.Fatal(msg, zap.Any("args", args))
}

// Implement Errorf method
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
	if len(fields)%2 != 0 {
		// Handle the case where fields are not in key-value pairs
		zapFields = append(zapFields, zap.String("error", "fields must be in key-value pairs"))
		return zapFields
	}
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			zapFields = append(zapFields, zap.Any("error", "key must be a string"))
			continue
		}
		zapFields = append(zapFields, zap.Any(key, fields[i+1]))
	}
	return zapFields
}
