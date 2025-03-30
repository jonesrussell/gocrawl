// Package logger provides logging functionality for the application.
package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger wraps zap.Logger to implement Interface
type ZapLogger struct {
	*zap.Logger
}

func (l *ZapLogger) Debug(msg string, fields ...any) {
	l.Logger.Debug(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Error(msg string, fields ...any) {
	l.Logger.Error(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Info(msg string, fields ...any) {
	l.Logger.Info(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Warn(msg string, fields ...any) {
	l.Logger.Warn(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Fatal(msg string, fields ...any) {
	l.Logger.Fatal(msg, zap.Any("fields", fields))
}

func (l *ZapLogger) Printf(format string, args ...any) {
	l.Logger.Sugar().Infof(format, args...)
}

func (l *ZapLogger) Errorf(format string, args ...any) {
	l.Logger.Sugar().Errorf(format, args...)
}

// Module provides the logger module for dependency injection.
var Module = fx.Module("logger",
	fx.Provide(
		// Provide the logger with configuration
		fx.Annotate(
			NewLogger,
			fx.As(new(Interface)),
		),
	),
)

// NewLogger creates a new logger instance with the given configuration.
func NewLogger(p Params) (*ZapLogger, error) {
	// Create the logger configuration
	config := zap.NewProductionConfig()
	if p.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Create the logger
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{Logger: logger}, nil
}
