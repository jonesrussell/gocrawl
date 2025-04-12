// Package logger provides logging functionality for the application.
package logger

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Params holds the parameters for creating a new logger.
type Params struct {
	Config *Config
	App    *app.Config
}

// Module provides the logger module.
var Module = fx.Module("logger",
	fx.Provide(New),
	fx.Provide(func(logger Interface) *zap.Logger {
		if l, ok := logger.(*Logger); ok {
			return l.zapLogger
		}
		return zap.NewNop()
	}),
)

// levelToZap converts a logger.Level to a zapcore.Level.
func levelToZap(level Level) zapcore.Level {
	switch string(level) {
	case string(DebugLevel):
		return zapcore.DebugLevel
	case string(InfoLevel):
		return zapcore.InfoLevel
	case string(WarnLevel):
		return zapcore.WarnLevel
	case string(ErrorLevel):
		return zapcore.ErrorLevel
	case string(FatalLevel):
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
