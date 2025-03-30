// Package job implements the job scheduler command.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/job"
	"go.uber.org/fx"
)

// Module provides the job command's dependencies.
var Module = fx.Module("job",
	fx.Provide(
		// Provide the done channel for job completion
		fx.Annotate(
			func() chan struct{} {
				return make(chan struct{})
			},
			fx.ResultTags(`name:"crawlDone"`),
		),
	),
	job.Module,
)
