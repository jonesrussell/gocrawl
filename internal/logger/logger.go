// Package logger provides logging functionality for the application.
package logger

import (
	"go.uber.org/zap"
)

// fieldPairSize represents the number of elements in a key-value pair.
const fieldPairSize = 2

// logger implements the Interface.
type logger struct {
	zapLogger *zap.Logger
	config    *Config
}

// Debug logs a debug message.
func (l *logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, toZapFields(fields)...)
}

// Info logs an info message.
func (l *logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, toZapFields(fields)...)
}

// Warn logs a warning message.
func (l *logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, toZapFields(fields)...)
}

// Error logs an error message.
func (l *logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, toZapFields(fields)...)
}

// Fatal logs a fatal message and exits.
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, toZapFields(fields)...)
}

// With creates a child logger with additional fields.
func (l *logger) With(fields ...any) Interface {
	return &logger{
		zapLogger: l.zapLogger.With(toZapFields(fields)...),
		config:    l.config,
	}
}

// toZapFields converts a list of any fields to zap.Field.
func toZapFields(fields []any) []zap.Field {
	if len(fields)%fieldPairSize != 0 {
		return []zap.Field{zap.Error(ErrInvalidFields)}
	}

	zapFields := make([]zap.Field, 0, len(fields)/fieldPairSize)
	for i := 0; i < len(fields); i += fieldPairSize {
		key, ok := fields[i].(string)
		if !ok {
			return []zap.Field{zap.Error(ErrInvalidFields)}
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
