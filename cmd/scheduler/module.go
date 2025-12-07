// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	internalcommon "github.com/jonesrussell/gocrawl/internal/common"
	"go.uber.org/fx"
)

// Module provides the scheduler command module for dependency injection.
// Note: Command registration is handled by Command() function, not through FX Group annotation.
// The scheduler command constructs dependencies directly, so no FX modules are needed here.
// The NewSchedulerService provider is kept for potential FX-based usage elsewhere.
var Module = fx.Module("scheduler",
	// Include required modules
	cmdcommon.Module,

	// Provide the scheduler service
	fx.Provide(
		// Provide the done channel
		func() chan struct{} {
			return make(chan struct{})
		},

		// Provide the active jobs counter
		func() *int32 {
			var jobs int32
			return &jobs
		},

		// Provide the scheduler service
		fx.Annotate(
			NewSchedulerService,
			fx.As(new(internalcommon.JobService)),
		),
	),
)
