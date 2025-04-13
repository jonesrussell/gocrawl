// Package job provides the job scheduler implementation.
package job

import (
	"go.uber.org/fx"
)

// Module provides the job module's dependencies.
var Module = fx.Module("job",
	fx.Provide(
		// Provide scheduler
		fx.Annotate(
			NewScheduler,
		),
	),
)
