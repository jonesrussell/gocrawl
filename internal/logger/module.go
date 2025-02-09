package logger

import (
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"go.uber.org/fx"
)

// Module provides the logger as an Fx module
var Module = fx.Module("logger",
	fx.Provide(
		NewLogger,
	),
)

// NewLogger initializes the appropriate logger based on the environment
func NewLogger(cfg *config.Config) (*CustomLogger, error) {
	env := cfg.AppEnv
	logLevel := cfg.LogLevel // Read log level from config

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
		level = INFO // Default to INFO if not set or invalid
	}

	if env == "development" {
		return NewDevelopmentLogger()
	}
	return NewCustomLogger(level) // Pass the log level from config
}

func InitializeLogger() (*CustomLogger, error) {
	return NewCustomLogger(INFO)
}
