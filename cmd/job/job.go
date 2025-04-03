// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/common"
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
	Sources sources.Interface
	Logger  logger.Interface

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Config holds the application configuration
	Config config.Interface

	// Context provides the context for the job scheduler
	Context context.Context

	// Processors is a slice of content processors, injected as a group
	Processors []common.Processor `group:"processors"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`

	// ActiveJobs tracks the number of currently running jobs
	ActiveJobs *int32 `optional:"true"`

	// Client is the Elasticsearch client
	Client *elasticsearch.Client
}

const (
	shutdownTimeout  = 30 * time.Second
	jobCheckInterval = 100 * time.Millisecond
)

// JobCommandDeps holds the dependencies for the job command
type JobCommandDeps struct {
	// Core dependencies
	Logger     logger.Interface
	Processors []common.Processor `group:"processors" json:"processors,omitempty"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(
	ctx context.Context,
	log logger.Interface,
	sources sources.Interface,
	c crawler.Interface,
	processors []common.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
	client *elasticsearch.Client,
) {
	log.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(jobCheckInterval)
	defer ticker.Stop()

	// Do initial check
	checkAndRunJobs(ctx, log, sources, c, time.Now(), processors, done, cfg, activeJobs, client)

	for {
		select {
		case <-ctx.Done():
			log.Info("Job scheduler shutting down")
			return
		case t := <-ticker.C:
			checkAndRunJobs(ctx, log, sources, c, t, processors, done, cfg, activeJobs, client)
		}
	}
}

// executeCrawl performs the crawl operation for a single source.
func executeCrawl(
	ctx context.Context,
	log logger.Interface,
	c crawler.Interface,
	source sources.Config,
	processors []common.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
	client *elasticsearch.Client,
) {
	atomic.AddInt32(activeJobs, 1)
	defer atomic.AddInt32(activeJobs, -1)

	collectorResult, err := app.SetupCollector(ctx, log, source, processors, done, cfg, client)
	if err != nil {
		log.Error("Error setting up collector",
			"error", err,
			"source", source.Name)
		return
	}

	if configErr := app.ConfigureCrawler(collectorResult, source); configErr != nil {
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
	log logger.Interface,
	sources sources.Interface,
	c crawler.Interface,
	now time.Time,
	processors []common.Processor,
	done chan struct{},
	cfg config.Interface,
	activeJobs *int32,
	client *elasticsearch.Client,
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

	for i := range sourcesList {
		source := &sourcesList[i]
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				log.Info("Running scheduled crawl",
					"source", source.Name,
					"time", scheduledTime)
				executeCrawl(ctx, log, c, *source, processors, done, cfg, activeJobs, client)
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

	// Start the job scheduler
	go runScheduler(
		p.Context,
		p.Logger,
		p.Sources,
		p.CrawlerInstance,
		p.Processors,
		p.Done,
		p.Config,
		p.ActiveJobs,
		p.Client,
	)

	return nil
}

// NewJobCommand creates a new job command with the given dependencies
func NewJobCommand(log logger.Interface) *cobra.Command {
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
	return fx.New(
		fx.Provide(
			fx.Annotate(
				context.Background,
				fx.ResultTags(`name:"jobContext"`),
			),
			fx.Annotate(
				func() *logger.Config {
					return &logger.Config{
						Level:       logger.InfoLevel,
						Development: false,
						Encoding:    "json",
					}
				},
				fx.ResultTags(`group:"loggerConfig"`),
			),
		),
		config.Module,
		logger.Module,
		sources.Module,
		crawler.Module,
		storage.Module,
		fx.Invoke(func(p Params) {
			if err := startJob(p); err != nil {
				p.Logger.Error("Error starting job scheduler", "error", err)
			}
		}),
	)
}
