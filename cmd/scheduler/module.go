// Package scheduler implements the job scheduler command for managing scheduled crawling tasks.
package scheduler

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the scheduler command module for dependency injection
var Module = fx.Module("scheduler",
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	sources.Module,
	crawler.Module,

	// Provide the context
	fx.Provide(context.Background),

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
			fx.As(new(common.JobService)),
		),
	),
)
