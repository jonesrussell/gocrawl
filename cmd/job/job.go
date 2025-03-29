// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync/atomic"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
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

const (
	shutdownTimeout  = 30 * time.Second
	jobCheckInterval = 100 * time.Millisecond
)

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

// startJob starts the job scheduler and waits for completion
func startJob(p Params) error {
	// Create a cancellable context tied to the command's context
	ctx, cancel := context.WithCancel(p.Context)
	defer cancel()

	// Create signal handler
	handler := signal.NewSignalHandler(p.Logger)
	handler.SetCleanup(func() {
		cancel()
	})

	// Set up signal handling
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Create done channel for shutdown coordination
	done := make(chan struct{})
	defer close(done)

	// Create configuration from Viper settings
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fxApp := fx.New(
		fx.Supply(cfg),
		fx.Supply(logger.NewNoOp()),
		fx.Supply(done),
		fx.Provide(
			NewJobScheduler,
			NewJobRunner,
		),
	)

	// Set the fx app for coordinated shutdown
	handler.SetFXApp(fxApp)

	// Start the application
	if err := fxApp.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for either context cancellation or job completion
	select {
	case <-ctx.Done():
		p.Logger.Info("Context cancelled, initiating shutdown")
	case <-done:
		p.Logger.Info("Job completed")
	}

	// Stop the application
	if err := fxApp.Stop(context.Background()); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}

	return nil
}

// JobCommandDeps holds the dependencies for the job command
type JobCommandDeps struct {
	// Core dependencies
	Logger     logger.Interface
	Config     config.Interface
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
}

// NewJobCommand creates a new job command with the given dependencies
func NewJobCommand(deps JobCommandDeps) *cobra.Command {
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
			handler := signal.NewSignalHandler(deps.Logger)
			cleanup := handler.Setup(ctx)
			defer cleanup()

			// Create a logger for command-level logging
			cmdLogger := deps.Logger

			// Initialize the Fx application
			var fxApp *fx.App
			var done chan struct{}
			var setupErr error
			fxApp, done, setupErr = setupFXApp(ctx, handler, cmdLogger)
			if setupErr != nil {
				return fmt.Errorf("failed to setup application: %w", setupErr)
			}

			// Set the fx app for coordinated shutdown
			handler.SetFXApp(fxApp)

			// Start the application
			if startErr := fxApp.Start(ctx); startErr != nil {
				return fmt.Errorf("failed to start application: %w", startErr)
			}

			// Wait for either context cancellation or job completion
			select {
			case <-ctx.Done():
				cmdLogger.Info("Context cancelled, initiating shutdown")
			case <-done:
				cmdLogger.Info("Job completed")
			}

			// Stop the application
			if stopErr := stopFXApp(fxApp); stopErr != nil {
				return fmt.Errorf("failed to stop application: %w", stopErr)
			}

			return nil
		},
	}

	return cmd
}

// Command returns the job command with default dependencies
func Command() *cobra.Command {
	// Create configuration from Viper settings
	cfg, err := config.New()
	if err != nil {
		// Return a command that will fail with the error
		return &cobra.Command{
			Use:   "job",
			Short: "Start the job scheduler",
			Long: "Start the job scheduler to manage and execute scheduled crawling tasks.\n" +
				"The scheduler will run continuously until interrupted with Ctrl+C.",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return fmt.Errorf("failed to create config: %w", err)
			},
		}
	}

	return NewJobCommand(JobCommandDeps{
		Logger: logger.NewNoOp(),
		Config: cfg,
	})
}

// Job represents a single crawling job
type Job struct {
	// Core dependencies
	Logger     logger.Interface
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
}

// JobScheduler manages the scheduling of crawling jobs
type JobScheduler struct {
	Logger     logger.Interface
	Config     config.Interface
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
}

// JobRunner executes crawling jobs
type JobRunner struct {
	Logger     logger.Interface
	Config     config.Interface
	Processors []collector.Processor `group:"processors" json:"processors,omitempty"`
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(
	logger logger.Interface,
	config config.Interface,
	processors []collector.Processor,
) *JobScheduler {
	return &JobScheduler{
		Logger:     logger,
		Config:     config,
		Processors: processors,
	}
}

// NewJobRunner creates a new job runner
func NewJobRunner(
	logger logger.Interface,
	config config.Interface,
	processors []collector.Processor,
) *JobRunner {
	return &JobRunner{
		Logger:     logger,
		Config:     config,
		Processors: processors,
	}
}

// setupFXApp initializes the Fx application with all required dependencies.
func setupFXApp(
	_ context.Context,
	_ *signal.SignalHandler,
	_ common.Logger,
) (*fx.App, chan struct{}, error) {
	done := make(chan struct{})

	// Create configuration from Viper settings
	cfg, err := config.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create config: %w", err)
	}

	fxApp := fx.New(
		fx.Supply(cfg),
		fx.Supply(logger.NewNoOp()),
		fx.Supply(done),
		fx.Provide(
			NewJobScheduler,
			NewJobRunner,
		),
	)

	return fxApp, done, nil
}

// stopFXApp stops the Fx application with a timeout.
func stopFXApp(fxApp *fx.App) error {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer stopCancel()

	if err := fxApp.Stop(stopCtx); err != nil {
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("error stopping application: %w", err)
		}
	}
	return nil
}
