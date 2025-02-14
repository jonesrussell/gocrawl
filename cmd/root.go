package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/app"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Initialize viper
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Read in the config file
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Create the fx app
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

		// Provide base context
		fx.Provide(func() context.Context {
			return context.Background()
		}),
	)

	ctx := context.Background()

	if err := appInstance.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Add commands with proper dependencies
	rootCmd.AddCommand(NewCrawlCmd(deps.Logger, deps.Crawler))
	rootCmd.AddCommand(NewSearchCmd(deps.Logger))

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	// Wait for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal is received or app is done
	select {
	case sig := <-sigChan:
		deps.Logger.Info("Received signal", "signal", sig)
	case <-ctx.Done():
		deps.Logger.Info("Context done")
	}

	// Shutdown the application
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
