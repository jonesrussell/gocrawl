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
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatalf(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
	Field(key string, value interface{}) zap.Field
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
func (c *CustomLogger) Info(msg string, fields ...zap.Field) {
	c.Logger.Info(msg, fields...)
}

// Error logs an error message
func (c *CustomLogger) Error(msg string, fields ...zap.Field) {
	c.Logger.Error(msg, fields...)
}

// Debug logs a debug message
func (c *CustomLogger) Debug(msg string, fields ...zap.Field) {
	c.Logger.Debug(msg, fields...)
}

// Warn logs a warning message
func (c *CustomLogger) Warn(msg string, fields ...zap.Field) {
	c.Logger.Warn(msg, fields...)
}

// Implement Fatalf method
func (z *CustomLogger) Fatalf(msg string, args ...interface{}) {
	z.Logger.Fatal(msg, zap.Any("args", args))
}

// Implement Errorf method
func (z *CustomLogger) Errorf(format string, args ...interface{}) {
	z.Logger.Error(fmt.Sprintf(format, args...))
}

// Field creates a zap.Field for structured logging
func (z *CustomLogger) Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
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
func (z *CustomLogger) GetZapLogger() *zap.Logger {
	return z.Logger
}
