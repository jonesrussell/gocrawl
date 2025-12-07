// Package logger provides logging functionality for the application.
package logger

import (
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the logger module.
var Module = fx.Module("logger",
	fx.Provide(
		// Provide the logger interface
		func(cfg *app.Config) (Interface, error) {
			// Determine log level - use debug if debug flag is set or in development
			level := InfoLevel
			if cfg.Debug || cfg.Environment == "development" {
				level = DebugLevel
			}

			// Create logger config with sensible defaults
			logConfig := &Config{
				Level:       level,
				Development: cfg.Environment != "production" || cfg.Debug,
				Encoding:    "console",
				EnableColor: true,
				OutputPaths: []string{"stdout"},
			}
			return New(logConfig)
		},
		// Provide the underlying zap logger for components that need it directly
		func(logger Interface) *zap.Logger {
			if l, ok := logger.(*Logger); ok {
				return l.zapLogger
			}
			return zap.NewNop()
		},
	),
)
