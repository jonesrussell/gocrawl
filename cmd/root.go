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

var cfgFile string

var globalLogger logger.Interface

var globalConfig *config.Config

var rootCmd = &cobra.Command{
	Use:   "gocrawl",
	Short: "A web crawler that stores content in Elasticsearch",
}

// Execute is the entry point for the CLI
func Execute() {
	// Initialize configuration
	var err error
	globalConfig, err = config.NewConfig() // This should be the only place you call NewConfig
	if err != nil {
		// Initialize logger before logging the error
		globalLogger, _ = InitializeLogger(&config.Config{}) // Initialize with an empty config to avoid nil logger
		globalLogger.Error("Error creating Config", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	globalLogger, err = InitializeLogger(globalConfig)
	if err != nil {
		globalLogger.Error("Error creating Logger", "error", err)
		os.Exit(1)
	}

	// Register the crawl and search commands
	rootCmd.AddCommand(NewCrawlCmd(globalLogger, globalConfig))  // Pass logger and config to crawl command
	rootCmd.AddCommand(NewSearchCmd(globalLogger, globalConfig)) // Pass logger and config to search command

	err = rootCmd.Execute()
	if err != nil {
		globalLogger.Error("Error executing root command", "error", err)
		os.Exit(1)
	}
}

// Initialize the command
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
}

// Initialize configuration
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
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

// Shutdown gracefully shuts down the application
func Shutdown(ctx context.Context) error {
	// Implement shutdown logic if necessary
	return nil
}
