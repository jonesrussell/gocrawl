// Package logger provides logging functionality for the application.
package logger

import (
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
				// Set log level based on config and debug flag
				var level Level = InfoLevel
				if params.Config != nil {
					level = params.Config.Level
				}
				if params.App != nil && (params.App.Debug || params.Config.Development) {
					level = DebugLevel
				}

				// Convert to zap level
				zapLevel := levelToZap(level)

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
				} else {
					encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
				}

				// Create core with appropriate encoder based on config
				var core zapcore.Core
				encoding := "console"
				if params.Config != nil {
					encoding = params.Config.Encoding
				}

				// Create the output
				output := zapcore.AddSync(os.Stdout)

				// Create level enabler
				levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= zapLevel
				})

				if encoding == "json" {
					core = zapcore.NewCore(
						zapcore.NewJSONEncoder(encoderConfig),
						output,
						levelEnabler,
					)
				} else {
					core = zapcore.NewCore(
						zapcore.NewConsoleEncoder(encoderConfig),
						output,
						levelEnabler,
					)
				}

				// Create logger
				zapLogger := zap.New(core,
					zap.AddCaller(),
					zap.AddStacktrace(zapcore.ErrorLevel),
				)

				// Log debug mode enablement if applicable
				if level == DebugLevel {
					zapLogger.Debug("Debug mode enabled",
						zap.Bool("app.Debug", params.App.Debug),
						zap.Bool("config.Development", params.Config.Development),
						zap.String("level", string(level)),
						zap.String("zapLevel", zapLevel.String()),
					)
				}

				return zapLogger, nil
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
