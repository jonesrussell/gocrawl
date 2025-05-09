// Package logger provides logging functionality for the application.
package logger

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ZapWrapper wraps a zap.Logger to implement logger.Interface.
type ZapWrapper struct {
	*zap.Logger
}

// Debug implements logger.Interface.
func (l *ZapWrapper) Debug(msg string, fields ...any) {
	l.Logger.Debug(msg, toZapFields(fields)...)
}

// Info implements logger.Interface.
func (l *ZapWrapper) Info(msg string, fields ...any) {
	l.Logger.Info(msg, toZapFields(fields)...)
}

// Error implements logger.Interface.
func (l *ZapWrapper) Error(msg string, fields ...any) {
	l.Logger.Error(msg, toZapFields(fields)...)
}

// Warn implements logger.Interface.
func (l *ZapWrapper) Warn(msg string, fields ...any) {
	l.Logger.Warn(msg, toZapFields(fields)...)
}

// Fatal implements logger.Interface.
func (l *ZapWrapper) Fatal(msg string, fields ...any) {
	l.Logger.Fatal(msg, toZapFields(fields)...)
}

// With implements logger.Interface.
func (l *ZapWrapper) With(fields ...any) Interface {
	return &ZapWrapper{
		Logger: l.Logger.With(toZapFields(fields)...),
	}
}

// NewLogger creates a new logger instance.
func NewLogger(cfg *app.Config) (Interface, error) {
	var config zap.Config
	if cfg.Environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapWrapper{Logger: logger}, nil
}

// Module provides the logger module.
var Module = fx.Module("logger",
	fx.Provide(
		NewLogger,
		func(logger Interface) *zap.Logger {
			if l, ok := logger.(*ZapWrapper); ok {
				return l.Logger
			}
			return zap.NewNop()
		},
	),
)
