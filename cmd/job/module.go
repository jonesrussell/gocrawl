// Package job implements the job scheduler command.
package job

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/job"
	"github.com/jonesrussell/gocrawl/pkg/logger"
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
		// Provide the logger
		fx.Annotate(
			func(p logger.Params) types.Logger {
				return logger.NewNoOp()
			},
			fx.As(new(types.Logger)),
		),
		// Provide the context
		fx.Annotate(
			func(lc fx.Lifecycle) context.Context {
				return context.Background()
			},
			fx.As(new(context.Context)),
		),
	),
	job.Module,
)
