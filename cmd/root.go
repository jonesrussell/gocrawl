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

var rootCmd = &cobra.Command{
	Use:   "gocrawl",
	Short: "A web crawler that stores content in Elasticsearch",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	var lgr *logger.CustomLogger // Declare a variable to hold the logger

	app := fx.New(
		// Core modules
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,

		// Application module
		app.Module,

		fx.Populate(&lgr), // Populate the logger from the Fx context
	)

	if err := app.Start(context.Background()); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(NewCrawlCmd(lgr)) // Pass logger from app

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	if err := app.Stop(context.Background()); err != nil {
		return fmt.Errorf("error stopping application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	app := fx.New(
		// Core modules
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,

		// Application module
		app.Module,
	)

	if err := app.Stop(ctx); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	return nil
}
