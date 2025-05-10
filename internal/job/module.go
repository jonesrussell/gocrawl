// Package job provides the job scheduler implementation.
package job

import (
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// SchedulerParams contains dependencies for creating a scheduler
type SchedulerParams struct {
	fx.In

	Logger  logger.Interface
	Sources *sources.Sources
	Storage storagetypes.Interface
}

// SchedulerResult contains the scheduler and its components
type SchedulerResult struct {
	fx.Out

	Scheduler Interface
}

// ProvideScheduler creates a new scheduler instance
func ProvideScheduler(p SchedulerParams) SchedulerResult {
	return SchedulerResult{
		Scheduler: NewScheduler(p.Logger, p.Sources, p.Storage),
	}
}

// Module provides the job module's dependencies.
var Module = fx.Module("job",
	fx.Provide(
		ProvideScheduler,
	),
)
