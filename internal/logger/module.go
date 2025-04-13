// Package logger provides logging functionality for the application.
package logger

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Params holds the parameters for creating a new logger.
type Params struct {
	Config *Config
	App    *app.Config
}

// Module provides the logger module.
var Module = fx.Module("logger",
	fx.Provide(func(logger Interface) *zap.Logger {
		if l, ok := logger.(*Logger); ok {
			return l.zapLogger
		}
		return zap.NewNop()
	}),
)
