// Package logger provides logging functionality for the application.
package logger

import (
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"go.uber.org/fx"
)

// Module provides the logger module and its dependencies using fx.
var Module = fx.Options(
	fx.Provide(
		func() (types.Logger, error) {
			return NewCustomLogger(nil, Params{
				Debug:  true,
				Level:  "debug",
				AppEnv: "development",
			})
		},
	),
)
