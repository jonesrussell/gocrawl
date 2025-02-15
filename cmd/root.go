package cmd

import (
	"context"
	"fmt"
	"os"

	// Import the Elasticsearch package
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		fmt.Printf("Error reading config file: %v\n", err)
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

func Execute() error {
	// Initialize config
	cfg, err := config.NewConfig(config.NewHTTPTransport())
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := InitializeLogger(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize logger: %v", err))
		os.Exit(1)
	}
	log.Debug("Logger initialized successfully")

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.URL},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
		APIKey:    cfg.Elasticsearch.APIKey,
	})
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create Elasticsearch client: %v", err))
		os.Exit(1)
	}

	// Add commands
	rootCmd.AddCommand(NewCrawlCmd(log, cfg, esClient))  // Pass logger, config, and esClient to crawl command
	rootCmd.AddCommand(NewSearchCmd(log, cfg, esClient)) // Pass logger, config, and esClient to search command
	log.Debug("Commands added to root command")

	return rootCmd.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(_ context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}
