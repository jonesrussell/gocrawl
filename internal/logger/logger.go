package logger

import (
	"go.uber.org/zap"
)

type LoggerInterface interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Field(key string, value interface{}) zap.Field
	// Add other methods as needed
}

type CustomLogger struct {
	logger *zap.Logger
}

func (z *CustomLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *CustomLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *CustomLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
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

// NewDevelopmentLogger initializes a new CustomLogger for development
func NewDevelopmentLogger() (*CustomLogger, error) {
	logger, err := zap.NewDevelopment() // Use development logger
	if err != nil {
		return nil, err
	}
	return &CustomLogger{logger: logger}, nil
}
