package logger

import (
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("logger",
	fx.Provide(
		NewLogger,
	),
)

// NewLogger initializes the appropriate logger based on the environment
func NewLogger(cfg *config.Config) (*CustomLogger, error) {
	logLevel := cfg.LogLevel
	var level zapcore.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	params := Params{
		Level:  level,
		AppEnv: cfg.AppEnv,
	}

	if cfg.AppEnv == "development" {
		return NewDevelopmentLogger(params)
	}
	return NewCustomLogger(params)
}

func InitializeLogger() (*CustomLogger, error) {
	params := Params{
		Level:  zapcore.InfoLevel,
		AppEnv: "production",
	}
	return NewCustomLogger(params)
}
