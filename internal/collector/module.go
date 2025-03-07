package collector

import (
	"go.uber.org/fx"
)

// Module provides the collector module and its dependencies
func Module() fx.Option {
	return fx.Module("collector",
		fx.Provide(
			New,
		),
	)
}
