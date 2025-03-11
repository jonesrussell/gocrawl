// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params holds the dependencies required for the job scheduler.
type Params struct {
	fx.In

	// Lifecycle manages the application's startup and shutdown hooks
	Lifecycle fx.Lifecycle

	// Sources provides access to configured crawl sources
	Sources *sources.Sources

	// CrawlerInstance handles the core crawling functionality
	CrawlerInstance crawler.Interface

	// Logger provides structured logging capabilities
	Logger logger.Interface

	// Config holds the application configuration
	Config config.Interface

	// Context provides the context for the job scheduler
	Context context.Context `name:"jobContext"`
}

// runScheduler manages the execution of scheduled jobs.
func runScheduler(ctx context.Context, log common.Logger, sources *sources.Sources, c crawler.Interface) {
	log.Info("Starting job scheduler")

	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Do initial check
	checkAndRunJobs(ctx, log, sources, c, time.Now())

	for {
		select {
		case <-ctx.Done():
			log.Info("Job scheduler shutting down")
			return
		case t := <-ticker.C:
			checkAndRunJobs(ctx, log, sources, c, t)
		}
	}
}

// checkAndRunJobs evaluates and executes scheduled jobs.
func checkAndRunJobs(
	ctx context.Context,
	log common.Logger,
	sources *sources.Sources,
	c crawler.Interface,
	now time.Time,
) {
	if sources == nil {
		log.Error("Sources configuration is nil")
		return
	}

	currentTime := now.Format("15:04")
	log.Info("Checking jobs", "current_time", currentTime)

	for _, source := range sources.Sources {
		for _, scheduledTime := range source.Time {
			if currentTime == scheduledTime {
				log.Info("Running scheduled crawl",
					"source", source.Name,
					"time", scheduledTime)

				// Configure crawler for this source
				c.SetMaxDepth(source.MaxDepth)
				if err := c.SetRateLimit(source.RateLimit); err != nil {
					log.Error("Error setting rate limit",
						"error", err,
						"source", source.Name)
					continue
				}

				// Start crawling
				if err := c.Start(ctx, source.URL); err != nil {
					log.Error("Error starting crawler",
						"error", err,
						"source", source.Name)
					continue
				}

				// Wait for crawl to complete
				c.Wait()
				log.Info("Crawl completed",
					"source", source.Name)
			}
		}
	}
}

// startJob initializes and starts the job scheduler.
func startJob(p Params) error {
	// Create cancellable context
	ctx, cancel := context.WithCancel(p.Context)
	defer cancel()

	// Start scheduler in background
	go runScheduler(ctx, p.Logger, p.Sources, p.CrawlerInstance)

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	p.Logger.Info("Job scheduler running. Press Ctrl+C to stop...")
	<-sigChan
	p.Logger.Info("Shutting down...")

	return nil
}

// provideSources creates a new Sources instance from the sources.yml file.
func provideSources() (*sources.Sources, error) {
	return sources.LoadFromFile("sources.yml")
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

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
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
				provideSources,
			),
			fx.Invoke(startJob),
		)

		// Start the application and handle any startup errors
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either:
		// - A signal interrupt (SIGINT/SIGTERM)
		// - Context cancellation
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
			cancel() // Cancel our context
		case <-ctx.Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
			return err
		}

		return nil
	},
}

// Command returns the job command.
func Command() *cobra.Command {
	return Cmd
}
