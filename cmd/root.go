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

var (
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
	}
	appInstance *fx.App
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	var lgr *logger.CustomLogger // Declare a variable to hold the logger

	appInstance = fx.New(
		// Core modules
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,

		// Application module
		app.Module,

		fx.Populate(&lgr), // Populate the logger from the Fx context
		fx.Provide(func() context.Context {
			return context.Background() // Provide a new context
		}),
	)

	// Create a context for the application and add the logger to it
	ctx := logger.WithContext(context.Background(), lgr.GetZapLogger()) // Use GetZapLogger() to get the zap.Logger

	if err := appInstance.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Add subcommands to the root command
	rootCmd.AddCommand(NewCrawlCmd(lgr)) // Pass logger from appInstance
	rootCmd.AddCommand(NewSearchCmd(lgr))

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
