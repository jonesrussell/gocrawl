package metrics

import "go.uber.org/fx"

// Module provides the metrics package as a dependency injection module.
var Module = fx.Options(
	fx.Provide(NewMetrics),
)
