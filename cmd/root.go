package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/multisource"
	"github.com/jonesrussell/gocrawl/internal/storage"
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

// Initialize configuration
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Failed to load configuration", err)
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
	sourceName := ""                                       // Initialize sourceName
	cmdInstance, err := initializeDependencies(sourceName) // Pass sourceName here
	if err != nil {
		fmt.Println("Failed to initialize dependencies", err)
		os.Exit(1)
	}

	return cmdInstance.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}

// initializeDependencies initializes all dependencies for the CLI
func initializeDependencies(sourceName string) (*cobra.Command, error) {
	// Initialize configuration
	cfg, err := config.NewConfig() // This should be the only place you call NewConfig
	if err != nil {
		return nil, err
	}

	// Initialize logger
	log, err := InitializeLogger(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Elasticsearch client using the provided function
	elasticClient, err := storage.ProvideElasticsearchClient(cfg, log) // Use the function from storage module
	if err != nil {
		return nil, err
	}

	// Initialize storage
	storageInstance, err := storage.NewStorage(elasticClient, log) // Ensure the storage is initialized
	if err != nil {
		return nil, err
	}

	// Initialize the debugger
	debuggerInstance := logger.NewCollyDebugger(log) // Pass the logger to the debugger

	// Initialize the crawler
	crawlerParams := crawler.Params{
		Logger:   log,
		Storage:  storageInstance,
		Debugger: debuggerInstance,
		Config:   cfg,
	}
	crawlerInstance, err := crawler.NewCrawler(crawlerParams) // Ensure the crawler is initialized
	if err != nil {
		return nil, err
	}

	// Initialize multisource
	multiSource, err := multisource.NewMultiSource(log, crawlerInstance.Crawler, "sources.yml", sourceName) // Pass logger, crawler, configPath, and sourceName
	if err != nil {
		return nil, err
	}

	// Create the multi crawl command
	multiCmd := NewMultiCrawlCmd(log, cfg, multiSource, crawlerInstance.Crawler) // Pass c here
	rootCmd.AddCommand(multiCmd)                                                 // Register the multi crawl command

	// Register the crawl and search commands
	rootCmd.AddCommand(NewCrawlCmd(log, cfg))  // Pass logger and config to crawl command
	rootCmd.AddCommand(NewSearchCmd(log, cfg)) // Pass logger and config to search command

	// Provide the sourceName for lifecycle hooks
	fx.Provide(func() string {
		return sourceName
	})

	return rootCmd, nil
}
