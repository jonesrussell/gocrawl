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
func Execute(cfg *config.Config) error {
	// Initialize the logger
	lgr, err := logger.NewLogger(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

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

	if err := app.Start(context.Background()); err != nil {
		lgr.Error("Error starting application", err)
		return err
	}

	if err := rootCmd.Execute(); err != nil {
		lgr.Error("Error executing root command", err)
		return err
	}

	if err := app.Stop(context.Background()); err != nil {
		lgr.Error("Error stopping application", err)
		return err
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
