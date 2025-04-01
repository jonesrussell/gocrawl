// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/pkg/app"
	"github.com/jonesrussell/gocrawl/pkg/collector"
	pkgconfig "github.com/jonesrussell/gocrawl/pkg/config"
	"github.com/jonesrussell/gocrawl/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params holds the dependencies required for the job operation.
type Params struct {
	fx.In
	Sources sources.Interface
	Logger  types.Logger

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Config holds the application configuration
	Config pkgconfig.Interface

	// Context provides the context for the job scheduler
	Context context.Context

	// Processors is a slice of content processors, injected as a group
	Processors []collector.Processor `group:"processors"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`

	// ActiveJobs tracks the number of currently running jobs
	ActiveJobs *int32 `optional:"true"`
}

const (
	shutdownTimeout  = 30 * time.Second
	jobCheckInterval = 100 * time.Millisecond
)

// JobCommandDeps holds the dependencies for the job command
type JobCommandDeps struct {
	// Core dependencies
	Logger     types.Logger
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(
	ctx context.Context,
	log types.Logger,
	sources sources.Interface,
	c crawler.Interface,
	processors []collector.Processor,
	done chan struct{},
	cfg pkgconfig.Interface,
	activeJobs *int32,
) {
	log.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(jobCheckInterval)
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
	log types.Logger,
	c crawler.Interface,
	source sources.Config,
	processors []collector.Processor,
	done chan struct{},
	cfg pkgconfig.Interface,
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
	log types.Logger,
	sources sources.Interface,
	c crawler.Interface,
	now time.Time,
	processors []collector.Processor,
	done chan struct{},
	cfg pkgconfig.Interface,
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

	sourcesList, err := sources.GetSources()
	if err != nil {
		log.Error("Failed to get sources", "error", err)
		return
	}

	for _, source := range sourcesList {
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
	if p.Sources == nil {
		return errors.New("sources configuration is required")
	}

	if p.CrawlerInstance == nil {
		return errors.New("crawler instance is required")
	}

	// Initialize active jobs counter if not provided
	if p.ActiveJobs == nil {
		var jobs int32
		p.ActiveJobs = &jobs
	}

	// Print loaded schedules
	p.Logger.Info("Loaded schedules:")
	sources, err := p.Sources.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	for _, source := range sources {
		if len(source.Time) > 0 {
			p.Logger.Info("Source schedule",
				"name", source.Name,
				"times", source.Time)
		}
	}

	// Start scheduler in background
	go runScheduler(p.Context, p.Logger, p.Sources, p.CrawlerInstance, p.Processors, p.Done, p.Config, p.ActiveJobs)

	p.Logger.Info("Job scheduler running. Press Ctrl+C to stop...")

	// Wait for crawler completion
	<-p.Done

	return nil
}

// NewJobCommand creates a new job command with the given dependencies
func NewJobCommand(log types.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Start the job scheduler",
		Long: "Start the job scheduler to manage and execute scheduled crawling tasks.\n" +
			"The scheduler will run continuously until interrupted with Ctrl+C.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create a cancellable context that's tied to the command's context
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Set up signal handling with a no-op logger initially
			handler := signalhandler.NewSignalHandler(logger.NewNoOp())
			cleanup := handler.Setup(ctx)
			defer cleanup()

			// Initialize the Fx application
			fxApp := setupFXApp()

			// Set the fx app for coordinated shutdown
			handler.SetFXApp(fxApp)

			// Start the application
			if startErr := fxApp.Start(ctx); startErr != nil {
				return fmt.Errorf("failed to start job scheduler: %w", startErr)
			}

			// Update the signal handler with the real logger
			handler.SetLogger(log)

			// Wait for shutdown signal
			if !handler.Wait() {
				return errors.New("job scheduler shutdown timeout or context cancellation")
			}

			return nil
		},
	}

	return cmd
}

// Command returns the job command with default dependencies
func Command() *cobra.Command {
	// Create a logger for command-level logging
	log := logger.NewNoOp()

	// Create the job command with dependencies
	return NewJobCommand(log)
}

// setupFXApp creates and configures the fx application with all required dependencies.
func setupFXApp() *fx.App {
	// Create the fx app with all required dependencies
	options := []fx.Option{
		fx.Supply(
			pkgconfig.Params{
				Environment: "development",
				Debug:       true,
				Command:     "job",
			},
			logger.Params{
				Debug:  true,
				Level:  "debug",
				AppEnv: "development",
			},
		),
		pkgconfig.Module, // Provide config first
		Module,           // Then job module
		crawler.Module,
		sources.Module,
		storage.Module,
		fx.NopLogger,
		fx.Invoke(func(p Params) {
			if err := startJob(p); err != nil {
				p.Logger.Error("Error starting job scheduler", "error", err)
			}
		}),
	}

	return fx.New(options...)
}

type Job struct {
	// Core dependencies
	Logger     types.Logger
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
	// ... existing code ...
}

// NewJob creates a new job instance.
func NewJob(
	logger types.Logger,
	processors []collector.Processor,
	// ... existing code ...
) *Job {
	return &Job{
		Logger:     logger,
		Processors: processors,
		// ... existing code ...
	}
}
