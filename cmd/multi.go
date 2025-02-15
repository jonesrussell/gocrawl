package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// NewMultiCrawlCmd creates a new command for crawling multiple sources
func NewMultiCrawlCmd(log logger.Interface) *cobra.Command {
	var multiCrawlCmd = &cobra.Command{
		Use:   "multi",
		Short: "Crawl multiple sources defined in sources.yml",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return setupMultiCrawlCmd(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeMultiCrawlCmd(cmd, log)
		},
	}

	return multiCrawlCmd
}

// setupMultiCrawlCmd handles the setup for the multi-crawl command
func setupMultiCrawlCmd(_ *cobra.Command) error {
	// You can set up any necessary configurations here if needed
	return nil
}

// executeMultiCrawlCmd handles the execution of the multi-crawl command
func executeMultiCrawlCmd(cmd *cobra.Command, log logger.Interface) error {
	// Initialize fx container
	app := newMultiCrawlFxApp()

	// Start the application
	if err := app.Start(cmd.Context()); err != nil {
		log.Error("Error starting application", "error", err)
		return fmt.Errorf("error starting application: %w", err)
	}
	defer func() {
		if err := app.Stop(cmd.Context()); err != nil {
			log.Error("Error stopping application", "error", err)
		}
	}()

	return nil
}

// newMultiCrawlFxApp initializes the Fx application with dependencies for multi-crawl
func newMultiCrawlFxApp() *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		multisource.Module,
		fx.Invoke(setupMultiLifecycleHooks),
	)
}

// setupMultiLifecycleHooks sets up the lifecycle hooks for the Fx application
func setupMultiLifecycleHooks(lc fx.Lifecycle, deps struct {
	fx.In
	Logger      logger.Interface
	MultiSource *multisource.MultiSource
}) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			deps.Logger.Debug("Starting multi-crawl application...")
			return deps.MultiSource.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			deps.Logger.Debug("Stopping multi-crawl application...")
			deps.MultiSource.Stop()
			return nil
		},
	})
}
