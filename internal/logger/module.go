package logger

import (
	"errors"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

// Module provides the logger as an Fx module
//
//nolint:gochecknoglobals // This is a module definition
var Module = fx.Module("logger",
	fx.Provide(NewLogger),
)

// NewLogger initializes the appropriate logger based on the environment
func NewLogger(cfg *config.Config) (*CustomLogger, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
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

// parseLogLevel parses the log level from the configuration
func parseLogLevel(logLevel string) (zapcore.Level, error) {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return zapcore.DebugLevel, nil
	case "INFO":
		return zapcore.InfoLevel, nil
	case "WARN":
		return zapcore.WarnLevel, nil
	case "ERROR":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, errors.New("invalid log level")
	}
}
