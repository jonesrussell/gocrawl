// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync/atomic"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params holds the dependencies required for the job operation.
type Params struct {
	fx.In
	Sources sources.Interface `json:"sources,omitempty"`
	Logger  common.Logger

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle `json:"lifecycle,omitempty"`

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface `json:"crawler_instance,omitempty"`

	// Config holds the application configuration
	Config config.Interface `json:"config,omitempty"`

	// Context provides the context for the job scheduler
	Context context.Context `name:"jobContext" json:"context,omitempty"`

	// Processors is a slice of content processors, injected as a group
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone" json:"done,omitempty"`

	// ActiveJobs tracks the number of currently running jobs
	ActiveJobs *int32 `optional:"true" json:"active_jobs,omitempty"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(
	ctx context.Context,
	log common.Logger,
	sources sources.Interface,
	c crawler.Interface,
	processors []collector.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
) {
	log.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Do initial check
	checkAndRunJobs(ctx, log, sources, c, time.Now(), processors, done, cfg, activeJobs)

	for {
		select {
		case <-ctx.Done():
			log.Info("Job scheduler shutting down")
			return
		case t := <-ticker.C:
			checkAndRunJobs(ctx, log, sources, c, t, processors, done, cfg, activeJobs)
		}
	}
}

// executeCrawl performs the crawl operation for a single source.
func executeCrawl(
	ctx context.Context,
	log common.Logger,
	c crawler.Interface,
	source sources.Config,
	processors []collector.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
) {
	atomic.AddInt32(activeJobs, 1)
	defer atomic.AddInt32(activeJobs, -1)

	collectorResult, err := app.SetupCollector(ctx, log, source, processors, done, cfg)
	if err != nil {
		log.Error("Error setting up collector",
			"error", err,
			"source", source.Name)
		return
	}

	if configErr := app.ConfigureCrawler(c, source, collectorResult); configErr != nil {
		log.Error("Error configuring crawler",
			"error", configErr,
			"source", source.Name)
		return
	}

	if startErr := c.Start(ctx, source.URL); startErr != nil {
		log.Error("Error starting crawler",
			"error", startErr,
			"source", source.Name)
		return
	}

	c.Wait()
	log.Info("Crawl completed", "source", source.Name)
}

// checkAndRunJobs evaluates and executes scheduled jobs.
func checkAndRunJobs(
	ctx context.Context,
	log common.Logger,
	sources sources.Interface,
	c crawler.Interface,
	now time.Time,
	processors []collector.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
) {
	if sources == nil {
		log.Error("Sources configuration is nil")
		return
	}

	if c == nil {
		log.Error("Crawler instance is nil")
		return
	}

	currentTime := now.Format("15:04")
	log.Info("Checking jobs", "current_time", currentTime)

	for _, source := range sources.GetSources() {
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				log.Info("Running scheduled crawl",
					"source", source.Name,
					"time", scheduledTime)
				executeCrawl(ctx, log, c, source, processors, done, cfg, activeJobs)
			}
		}
	}
}

// startJob initializes and starts the job scheduler.
func startJob(p Params) error {
	// Create cancellable context
	ctx, cancel := context.WithCancel(p.Context)
	defer cancel()

	// Initialize active jobs counter if not provided
	if p.ActiveJobs == nil {
		var jobs int32
		p.ActiveJobs = &jobs
	}

	// Print loaded schedules
	p.Logger.Info("Loaded schedules:")
	for _, source := range p.Sources.GetSources() {
		if len(source.Time) > 0 {
			p.Logger.Info("Source schedule",
				"name", source.Name,
				"times", source.Time)
		}
	}

	// Start scheduler in background
	go runScheduler(ctx, p.Logger, p.Sources, p.CrawlerInstance, p.Processors, p.Done, p.Config, p.ActiveJobs)

	p.Logger.Info("Job scheduler running. Press Ctrl+C to stop...")

	// Wait for crawler completion
	<-p.Done

	return nil
}

// Command returns the job command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Start the job scheduler",
		Long: `Start the job scheduler to manage and execute scheduled crawling tasks.
The scheduler will run continuously until interrupted with Ctrl+C.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create a cancellable context
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Set up signal handling with a no-op logger initially
			handler := signal.NewSignalHandler(logger.NewNoOp())
			handler.Setup(ctx)

			// Initialize the Fx application
			fxApp := fx.New(
				fx.NopLogger,
				fx.Supply(
					fx.Annotate(ctx, fx.As(new(context.Context))),
					fx.Annotate(ctx, fx.As(new(context.Context)), fx.ResultTags(`name:"jobContext"`)),
					fx.Annotate(make(chan struct{}), fx.ResultTags(`name:"crawlDone"`)),
				),
				fx.Provide(
					func(cfg config.Interface) (string, error) {
						sources := cfg.GetSources()
						if len(sources) == 0 {
							return "", errors.New("no sources configured")
						}
						return sources[0].Name, nil
					},
					fx.Annotate(
						func(cfg config.Interface) string {
							sources := cfg.GetSources()
							if len(sources) == 0 {
								return ""
							}
							return sources[0].Name
						},
						fx.ResultTags(`name:"sourceName"`),
					),
				),
				common.Module,
				api.Module,
				storage.Module,
				crawler.Module,
				collector.Module,
				fx.Invoke(func(p Params) {
					// Update the signal handler with the real logger
					handler.SetLogger(p.Logger)
					if err := startJob(p); err != nil {
						p.Logger.Error("Error starting job scheduler", "error", err)
					}
					// Wait for shutdown signal
					handler.Wait()
				}),
			)

			// Set the fx app for coordinated shutdown
			handler.SetFXApp(fxApp)

			// Start the application
			if err := fxApp.Start(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("error starting application: %w", err)
			}

			return nil
		},
	}

	return cmd
}

type Job struct {
	// Core dependencies
	Logger     logger.Interface
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
	// ... existing code ...
}

// NewJob creates a new job instance.
func NewJob(
	logger logger.Interface,
	processors []collector.Processor,
	// ... existing code ...
) *Job {
	return &Job{
		Logger:     logger,
		Processors: processors,
		// ... existing code ...
	}
}
