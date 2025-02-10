package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector as an Fx module
//
//nolint:gochecknoglobals // This is a module
var Module = fx.Module("collector",
	fx.Provide(
		New,
	),
)
