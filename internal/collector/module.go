package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector as an Fx module
var Module = fx.Module("collector",
	fx.Provide(
		New,
	),
)
