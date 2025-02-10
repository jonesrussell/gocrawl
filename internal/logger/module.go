package logger

import (
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
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
	var level LogLevel
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	default:
		level = INFO
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
		Level:  INFO,
		AppEnv: "production", // default to production environment
	}
	return NewCustomLogger(params)
}
