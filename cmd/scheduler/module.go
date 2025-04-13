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
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// Module provides the scheduler command module for dependency injection
var Module = fx.Options(
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	sources.Module,
	crawler.Module,

	// Provide the context
	fx.Provide(context.Background),

	// Provide the done channel
	fx.Provide(func() chan struct{} {
		return make(chan struct{})
	}),

	// Provide the active jobs counter
	fx.Provide(func() *int32 {
		var jobs int32
		return &jobs
	}),

	// Provide the scheduler service
	fx.Provide(fx.Annotate(
		func(
			logger logger.Interface,
			storage types.Interface,
			sources sources.Interface,
			crawler crawler.Interface,
			done chan struct{},
			config config.Interface,
			processorFactory crawler.ProcessorFactory,
		) common.JobService {
			return NewSchedulerService(SchedulerServiceParams{
				Logger:           logger,
				Sources:          sources,
				Crawler:          crawler,
				Done:             done,
				Config:           config,
				Storage:          storage,
				ProcessorFactory: processorFactory,
			})
		},
		fx.As(new(common.JobService)),
	)),
)
