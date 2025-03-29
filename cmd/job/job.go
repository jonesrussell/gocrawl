// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
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

			// Create a logger for command-level logging
			cmdLogger := deps.Logger

			// Initialize the Fx application
			fxApp, done := setupFXApp(ctx, cmdLogger, deps.Config)

			// Set up signal handling with the logger
			handler := signal.NewSignalHandler(cmdLogger)
			handler.SetFXApp(fxApp)
			cleanup := handler.Setup(ctx)
			defer cleanup()

			// Start the application
			if startErr := fxApp.Start(ctx); startErr != nil {
				return fmt.Errorf("failed to start application: %w", startErr)
			}

			// Wait for either context cancellation or job completion
			select {
			case <-ctx.Done():
				cmdLogger.Info("Context cancelled, initiating shutdown...")
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

	// Create a logger for command-level logging
	log := logger.NewNoOp()

	return NewJobCommand(JobCommandDeps{
		Logger: log,
		Config: cfg,
	})
}

// setupFXApp initializes the Fx application with all required dependencies.
func setupFXApp(
	ctx context.Context,
	logger common.Logger,
	config config.Interface,
) (*fx.App, chan struct{}) {
	done := make(chan struct{})

	// Create the fx app with all required dependencies
	options := []fx.Option{
		fx.Supply(logger),
		fx.Supply(config),
		fx.Supply(done),
		fx.Supply(ctx),
		Module,
		fx.NopLogger,
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						// Simulate job completion after a short delay
						time.Sleep(100 * time.Millisecond)
						close(done)
					}()
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	}

	fxApp := fx.New(options...)

	return fxApp, done
}

// stopFXApp stops the Fx application with a timeout.
func stopFXApp(fxApp *fx.App) error {
	stopCtx, stopCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer stopCancel()

	if err := fxApp.Stop(stopCtx); err != nil {
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to stop application: %w", err)
		}
	}
	return nil
}
