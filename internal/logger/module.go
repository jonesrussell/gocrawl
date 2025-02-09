package logger

import (
	"os"

	"go.uber.org/fx"
)

// Module provides the logger as an Fx module
var Module = fx.Module("logger",
	fx.Provide(
		NewLogger,
	),
)

// NewLogger initializes the appropriate logger based on the environment
func NewLogger() (*CustomLogger, error) {
	env := os.Getenv("APP_ENV")
	if env == "development" {
		return NewDevelopmentLogger()
	}
	return NewCustomLogger()
}
