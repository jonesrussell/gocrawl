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

type ZapLogger struct {
	logger *zap.Logger
}

func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

// Field creates a zap.Field for structured logging
func (z *ZapLogger) Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// NewLogger initializes a new ZapLogger
func NewLogger() (LoggerInterface, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger: logger}, nil
}
