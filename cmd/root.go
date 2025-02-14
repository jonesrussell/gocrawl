package cmd

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// CrawlerDeps holds the dependencies needed for the crawler
type CrawlerDeps struct {
	fx.In

	Crawler *crawler.Crawler
	Logger  *logger.CustomLogger
}

var (
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
	}
	appInstance *fx.App
	deps        CrawlerDeps
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	appInstance = fx.New(
		// Core modules
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,

		// Application module
		app.Module,

		fx.Populate(&deps),

		fx.Provide(func() context.Context {
			return context.Background()
		}),
	)

	// Create a context for the application and add the logger to it
	ctx := context.Background()

	if err := appInstance.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(NewCrawlCmd(deps.Logger, deps.Crawler))
	rootCmd.AddCommand(NewSearchCmd(deps.Logger))

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	if err := appInstance.Stop(ctx); err != nil {
		return fmt.Errorf("error stopping application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	if err := appInstance.Stop(ctx); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	return nil
}
