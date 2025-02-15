package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
	}
)

// Initialize the command
func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		//nolint:forbidigo // This is a CLI error
		fmt.Println("Failed to load configuration:", err)
		os.Exit(1)
	}
}

// InitializeLogger initializes the logger
func InitializeLogger(cfg *config.Config) (logger.Interface, error) {
	env := cfg.App.Environment
	if env == "" {
		env = "development" // Set a default environment
	}

	if env == "development" {
		return logger.NewDevelopmentLogger() // Use colored logger in development
	}
	return logger.NewProductionLogger(cfg) // Use a different logger for production
}

// Execute is the entry point for the CLI
func Execute() error {
	// Initialize dependencies
	rootCmd, err := initializeDependencies()
	if err != nil {
		//nolint:forbidigo // This is a CLI error
		fmt.Println("Failed to initialize dependencies:", err)
		os.Exit(1)
	}

	return rootCmd.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(_ context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}

// initializeDependencies initializes all dependencies for the CLI
func initializeDependencies() (*cobra.Command, error) {
	// Initialize configuration
	cfg, err := config.NewConfig() // Example config initialization
	if err != nil {
		return nil, err
	}

	// Initialize logger
	log, err := InitializeLogger(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize fx app with all necessary modules
	app := fx.New(
		fx.Provide(
			func() *config.Config {
				return cfg
			},
			func() logger.Interface {
				return log
			},
		),
		config.Module,
		logger.Module,
		crawler.Module,
		multisource.Module,
	)

	// Create the multi crawl command
	multiCmd := NewMultiCrawlCmd(log, cfg, nil) // MultiSource will be resolved by Fx

	return multiCmd, app.Start(context.Background())
}
