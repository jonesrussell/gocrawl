package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}
}

// InitializeLogger initializes the logger
func InitializeLogger() (logger.Interface, error) {
	env := viper.GetString("APP_ENV")
	if env == "" {
		env = "development" // Set a default environment
	}

	if env == "development" {
		return logger.NewDevelopmentLogger() // Use colored logger in development
	}
	return logger.NewProductionLogger() // Use a different logger for production
}

func Execute() error {
	// Initialize logger
	log, err := InitializeLogger()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize logger: %v", err))
		os.Exit(1)
	}
	log.Debug("Logger initialized successfully")

	// Initialize configuration
	transport := config.NewHTTPTransport()  // Create the transport
	cfg, err := config.NewConfig(transport) // Pass the transport to NewConfig
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize config: %v", err))
		os.Exit(1)
	}

	// Initialize storage
	esClient, err := storage.ProvideElasticsearchClient()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize Elasticsearch client: %v", err))
		os.Exit(1)
	}

	storageInstance, err := storage.NewStorage(esClient)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize storage: %v", err))
		os.Exit(1)
	}

	// Add commands
	rootCmd.AddCommand(NewCrawlCmd(log, cfg, storageInstance)) // Pass logger, config, and storage to crawl command
	rootCmd.AddCommand(NewSearchCmd(log))                      // Pass logger to search command
	log.Debug("Commands added to root command")

	return rootCmd.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(_ context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}
