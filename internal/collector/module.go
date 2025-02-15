package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector module and its dependencies
var Module = fx.Module("collector",
	fx.Provide(
		New, // Function to create a new collector instance
	),
)
