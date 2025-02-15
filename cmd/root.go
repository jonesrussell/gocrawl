package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sourceFile string // Variable to hold the source file flag

	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if sourceFile != "" {
				viper.Set("CRAWLER_SOURCE_FILE", sourceFile) // Set the source file in viper
			}
		},
	}
)

// Initialize the command
func init() {
	cobra.OnInitialize(initConfig)

	// Add the source file flag
	rootCmd.PersistentFlags().StringVar(&sourceFile, "source", "", "Path to the source configuration file")
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

func Execute() error {
	// Initialize config
	cfg, err := config.NewConfig()
	if err != nil {
		//nolint:forbidigo // This is a CLI error
		fmt.Println("Failed to load configuration:", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := InitializeLogger(cfg)
	if err != nil {
		//nolint:forbidigo // This is a CLI error
		fmt.Println("Failed to initialize logger:", err)
		os.Exit(1)
	}
	log.Debug("Logger initialized successfully")

	// Add commands
	rootCmd.AddCommand(NewCrawlCmd(log, cfg))  // Pass logger and config to crawl command
	rootCmd.AddCommand(NewSearchCmd(log, cfg)) // Pass logger and config to search command
	log.Debug("Commands added to root command")

	return rootCmd.Execute()
}

// Shutdown gracefully shuts down the application
func Shutdown(_ context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}
