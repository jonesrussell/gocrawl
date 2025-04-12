// Package logger provides logging functionality for the application.
package logger

import (
	"fmt"
	"os"

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
	fx.Provide(
		fx.Annotate(
			func(params Params) (*zap.Logger, error) {
				// Create encoder config first
				encoderConfig := zapcore.EncoderConfig{
					TimeKey:        "ts",
					LevelKey:       "level",
					NameKey:        "logger",
					CallerKey:      "caller",
					FunctionKey:    zapcore.OmitKey,
					MessageKey:     "msg",
					StacktraceKey:  "stacktrace",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}

				// Check if we should enable color
				enableColor := false
				if params.Config != nil {
					enableColor = params.Config.EnableColor
				}
				if enableColor || (params.App != nil && params.App.Debug) {
					encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
					fmt.Fprintf(os.Stderr, "Color output enabled\n")
				} else {
					encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
					fmt.Fprintf(os.Stderr, "Color output disabled\n")
				}

				// Set log level based on config and debug flag
				level := zapcore.InfoLevel
				if params.Config != nil {
					level = levelToZap(params.Config.Level)
				}
				if params.App != nil && params.App.Debug {
					level = zapcore.DebugLevel
					fmt.Fprintf(os.Stderr, "Debug mode enabled, setting log level to DEBUG\n")
				}

				// Create core with appropriate encoder based on config
				var core zapcore.Core
				encoding := "console"
				if params.Config != nil {
					encoding = params.Config.Encoding
				}

				if encoding == "json" {
					core = zapcore.NewCore(
						zapcore.NewJSONEncoder(encoderConfig),
						zapcore.AddSync(os.Stdout),
						level,
					)
				} else {
					core = zapcore.NewCore(
						zapcore.NewConsoleEncoder(encoderConfig),
						zapcore.AddSync(os.Stdout),
						level,
					)
				}

				// Create logger with caller info enabled
				return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
			},
			fx.ResultTags(`name:"zapLogger"`),
		),
		fx.Annotate(
			func(zapLogger *zap.Logger) Interface {
				return &logger{
					zapLogger: zapLogger,
				}
			},
			fx.ResultTags(`name:"logger"`),
		),
	),
)

// levelToZap converts a logger.Level to a zapcore.Level.
func levelToZap(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
