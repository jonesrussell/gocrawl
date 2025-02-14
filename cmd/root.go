package cmd

import (
	"context"
	"fmt"
	"os"

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Let the command's PreRunE run first to update viper values
			if cmd.PreRunE != nil {
				if err := cmd.PreRunE(cmd, args); err != nil {
					return err
				}
			}

			// Now create the fx app with the updated viper values
			appInstance = fx.New(
				// Core modules
				config.Module,
				logger.Module,
				storage.Module,
				collector.Module,
				crawler.Module,
				app.Module,

				fx.Populate(&deps),
				fx.Provide(func() context.Context {
					return context.Background()
				}),
			)

			// Start the application
			if err := appInstance.Start(cmd.Context()); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if err := appInstance.Stop(cmd.Context()); err != nil {
				return fmt.Errorf("error stopping application: %w", err)
			}
			return nil
		},
	}
	appInstance *fx.App
	deps        CrawlerDeps
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}
}

func Execute() error {
	rootCmd.AddCommand(NewCrawlCmd(deps.Logger, deps.Crawler))
	rootCmd.AddCommand(NewSearchCmd(deps.Logger))

	return rootCmd.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	if err := appInstance.Stop(ctx); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	return nil
}
