package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerInterface interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatalf(msg string, args ...interface{})
	Field(key string, value interface{}) zap.Field
	// Add other methods as needed
}

type CustomLogger struct {
	logger *zap.Logger
}

// Ensure CustomLogger implements LoggerInterface
var _ LoggerInterface = (*CustomLogger)(nil)

func (z *CustomLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *CustomLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *CustomLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

func (z *CustomLogger) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

// Implement Fatalf method
func (z *CustomLogger) Fatalf(msg string, args ...interface{}) {
	z.logger.Fatal(msg, zap.Any("args", args)) // Log the fatal message, Zap will handle the exit
}

// Field creates a zap.Field for structured logging
func (z *CustomLogger) Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// NewCustomLogger initializes a new CustomLogger
func NewCustomLogger() (*CustomLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &CustomLogger{logger: logger}, nil
}

// NewDevelopmentLogger initializes a new CustomLogger for development with color
func NewDevelopmentLogger() (*CustomLogger, error) {
	// Create a new console encoder with color support
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Use colored level encoding

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		zap.DebugLevel, // Set the log level to Debug for development
	)

	logger := zap.New(core)
	return &CustomLogger{logger: logger}, nil
}

// Sync flushes any buffered log entries
func (z *CustomLogger) Sync() error {
	return z.logger.Sync()
}
