// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"fmt"
	"time"

	"errors"
	"sync/atomic"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/app"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// shutdownTimeoutSeconds is the number of seconds to wait for jobs to complete during shutdown
	shutdownTimeoutSeconds = 30
)

// Params holds the dependencies required for the job scheduler.
type Params struct {
	fx.In

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// Sources provides access to configured crawl sources
	Sources sources.Interface `name:"sourceManager"`

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Logger provides structured logging capabilities
	Logger logger.Interface

	// Config holds the application configuration
	Config config.Interface

	// Context provides the context for the job scheduler
	Context context.Context `name:"jobContext"`

	// Processors is a slice of content processors, injected as a group
	Processors []models.ContentProcessor `group:"processors"`

	// Done is a channel that signals when the crawl operation is complete
	Done chan struct{} `name:"crawlDone"`

	// ActiveJobs tracks the number of currently running jobs
	ActiveJobs *int32 `optional:"true"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(
	ctx context.Context,
	log common.Logger,
	sources sources.Interface,
	c crawler.Interface,
	processors []models.ContentProcessor,
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
	processors []models.ContentProcessor,
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
	processors []models.ContentProcessor,
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

	// Set up signal handling
	handler := signal.NewSignalHandler()
	cleanup := handler.Setup(ctx)
	defer cleanup()

	p.Logger.Info("Job scheduler running. Press Ctrl+C to stop...")

	// Wait for shutdown signal
	handler.Wait()

	// Log graceful shutdown
	p.Logger.Info("") // Add newline before shutdown messages
	p.Logger.Info("Starting graceful shutdown...")

	if atomic.LoadInt32(p.ActiveJobs) > 0 {
		p.Logger.Info("Waiting for in-progress jobs to complete...")
		common.PrintInfof("Press Ctrl+C again to force immediate shutdown")

		// Give time for any in-progress jobs to complete
		timeout := time.Duration(shutdownTimeoutSeconds) * time.Second
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
		defer shutdownCancel()

		// Wait for any in-progress jobs
		select {
		case <-p.Done:
			p.Logger.Info("All jobs completed successfully")
		case <-shutdownCtx.Done():
			p.Logger.Warn("Some jobs may not have completed - graceful shutdown timeout reached")
		}
	} else {
		p.Logger.Info("No active jobs - shutting down immediately")
	}

	p.Logger.Info("Job scheduler stopped")
	return nil
}

// Cmd represents the job scheduler command.
var Cmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl jobs",
	Long:  `Schedule and run crawl jobs based on the times specified in sources.yml`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Create a parent context that can be cancelled
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Initialize the Fx application with required modules and dependencies
		fxApp := fx.New(
			fx.NopLogger, // Suppress Fx startup/shutdown logs
			// Core dependencies
			config.Module,
			sources.Module,
			storage.Module,
			logger.Module,

			// Feature modules
			article.Module,
			content.Module,
			collector.Module(),
			crawler.Module,

			fx.Provide(
				fx.Annotate(
					func() context.Context {
						return ctx
					},
					fx.ResultTags(`name:"jobContext"`),
				),
				func() *int32 {
					var jobs int32
					return &jobs
				},
				fx.Annotate(
					func(sources sources.Interface) (string, string) {
						// For job scheduler, we'll use the first source's indices
						if len(sources.GetSources()) > 0 {
							allSources := sources.GetSources()
							return allSources[0].Index, allSources[0].ArticleIndex
						}
						return "content", "articles" // Default indices if no sources
					},
					fx.ParamTags(`name:"sourceManager"`),
					fx.ResultTags(`name:"contentIndex"`, `name:"indexName"`),
				),
				fx.Annotate(
					func(sources sources.Interface) string {
						// For job scheduler, we'll use the first source's name
						if len(sources.GetSources()) > 0 {
							return sources.GetSources()[0].Name
						}
						return "default" // Default source name if no sources
					},
					fx.ParamTags(`name:"sourceManager"`),
					fx.ResultTags(`name:"sourceName"`),
				),
				fx.Annotate(
					func() chan struct{} {
						return make(chan struct{})
					},
					fx.ResultTags(`name:"crawlDone"`),
				),
				func() chan *models.Article {
					return make(chan *models.Article, app.DefaultChannelBufferSize)
				},
			),
			fx.Invoke(startJob),
		)

		// Start the application
		if err := fxApp.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Set up signal handling
		handler := signal.NewSignalHandler()
		cleanup := handler.Setup(ctx)
		defer cleanup()

		// Set up cleanup for graceful shutdown
		handler.SetCleanup(func() {
			// Create a context with timeout for graceful shutdown
			stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
			defer stopCancel()

			// Stop the application and handle any shutdown errors
			if err := fxApp.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
				common.PrintErrorf("Error during shutdown: %v", err)
			}
		})

		// Wait for shutdown signal
		handler.Wait()

		return nil
	},
}

// Command returns the job command.
func Command() *cobra.Command {
	return Cmd
}
