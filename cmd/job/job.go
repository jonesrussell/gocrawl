// Package job implements the job scheduler command for managing scheduled crawling tasks.
package job

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params defines the parameters for job execution.
type Params struct {
	fx.In

	// Logger provides structured logging capabilities
	Logger common.Logger

	// Sources provides access to source configuration
	Sources *sources.Sources `optional:"true"`

	// Storage provides data persistence operations
	Storage common.Storage `optional:"true"`
}

// startJobScheduler initializes and manages the job scheduler lifecycle.
// It:
// - Creates a cancellable context for managing scheduled jobs
// - Starts the scheduler in a goroutine
// - Handles graceful shutdown when the application stops
func startJobScheduler(p Params, rootCmd string) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle) {
		ctx, cancel := context.WithCancel(context.Background())

		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				go runScheduler(ctx, p, rootCmd)
				return nil
			},
			OnStop: func(_ context.Context) error {
				cancel()
				return nil
			},
		})
	})
}

// runScheduler manages the execution of scheduled jobs.
// It:
// - Runs a ticker to check for scheduled jobs every minute
// - Handles context cancellation for graceful shutdown
// - Delegates job execution to checkAndRunJobs
func runScheduler(ctx context.Context, p Params, rootCmd string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.Logger.Info("Job scheduler shutting down")
			return
		case now := <-ticker.C:
			checkAndRunJobs(p, rootCmd, now)
		}
	}
}

// checkAndRunJobs evaluates and executes scheduled jobs.
// It:
// - Checks each source's configured schedule times
// - Executes crawl commands when scheduled times match
// - Logs job execution and any errors
func checkAndRunJobs(p Params, rootCmd string, now time.Time) {
	for _, source := range p.Sources.Sources {
		for _, t := range source.Time {
			scheduledTime, parseErr := time.Parse("15:04", t)
			if parseErr != nil {
				p.Logger.Error("Error parsing time", "error", parseErr, "source", source.Name, "time", t)
				continue
			}

			p.Logger.Debug("Checking scheduled time",
				"source", source.Name,
				"current_time", now.Format("15:04"),
				"scheduled_time", t,
				"current_hour", now.Hour(),
				"scheduled_hour", scheduledTime.Hour(),
				"current_minute", now.Minute(),
				"scheduled_minute", scheduledTime.Minute(),
			)

			if now.Hour() == scheduledTime.Hour() && now.Minute() == scheduledTime.Minute() {
				p.Logger.Info("Running scheduled crawl",
					"source", source.Name,
					"time", t,
					"current_time", now.Format("15:04"),
				)

				if err := runCrawlCommand(rootCmd, source.Name, p.Logger); err != nil {
					p.Logger.Error("Error running crawl command",
						"error", err,
						"source", source.Name,
						"time", t,
					)
				}
			}
		}
	}
}

// runCrawlCommand executes a crawl command for a specific source.
// It:
// - Constructs the command with appropriate arguments
// - Sets up stdout and stderr
// - Logs command execution
func runCrawlCommand(rootCmd, sourceName string, logger common.Logger) error {
	cmdArgs := []string{"crawl", sourceName}
	cmd := exec.Command(rootCmd, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Debug("Executing crawl command", "command", rootCmd, "args", cmdArgs)
	return cmd.Run()
}

// Cmd represents the job scheduler command that manages scheduled crawling tasks.
var Cmd = &cobra.Command{
	Use:   "job",
	Short: "Schedule and run crawl jobs",
	Long:  `Schedule and run crawl jobs based on the times specified in sources.yml`,
	Run: func(cmd *cobra.Command, _ []string) {
		rootPath := cmd.Root().Name()

		// Initialize the Fx application with required modules and dependencies
		app := fx.New(
			common.Module,
			Module,
			fx.Provide(
				func() *cobra.Command { return cmd },
			),
			fx.Invoke(func(p Params) {
				startJobScheduler(p, rootPath)
			}),
		)

		// Start the application and handle any startup errors
		if err := app.Start(cmd.Context()); err != nil {
			common.PrintErrorf("Error starting application: %v", err)
			os.Exit(1)
		}

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultShutdownTimeout)
		defer func() {
			cancel()
			if err := app.Stop(ctx); err != nil {
				common.PrintErrorf("Error during shutdown: %v", err)
				os.Exit(1)
			}
		}()
	},
}

// Command returns the job command for use in the root command
func Command() *cobra.Command {
	return Cmd
}
