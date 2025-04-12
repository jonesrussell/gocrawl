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

				// Always enable color in development mode
				encoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
					var color string
					switch l {
					case zapcore.DebugLevel:
						color = ColorDebug
					case zapcore.InfoLevel:
						color = ColorInfo
					case zapcore.WarnLevel:
						color = ColorWarn
					case zapcore.ErrorLevel:
						color = ColorError
					case zapcore.DPanicLevel:
						color = ColorError
					case zapcore.PanicLevel:
						color = ColorError
					case zapcore.FatalLevel:
						color = ColorFatal
					case zapcore.InvalidLevel:
						color = ColorReset
					default:
						color = ColorReset
					}
					enc.AppendString(color + l.CapitalString() + ColorReset)
				}

				// Create core with debug level if debug flag is set
				var core zapcore.Core
				level := zapcore.InfoLevel
				if params.App != nil && params.App.Debug {
					level = zapcore.DebugLevel
					fmt.Fprintf(os.Stderr, "Debug mode enabled, setting log level to DEBUG\n")
				}

				if config := params.Config; config != nil && config.Encoding == "json" {
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
				return zap.New(core, zap.AddCaller()), nil
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
