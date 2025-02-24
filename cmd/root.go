package cmd

import (
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
	fmt.Println("Executing command: Execute")

	// Register the crawl and search commands
	if err := rootCmd.Execute(); err != nil {
		globalLogger.Error("Error executing root command", "error", err)
		os.Exit(1)
	}
}

// Initialize the command
func init() {
	fmt.Println("Initializing command: init")
	cobra.OnInitialize(initConfig, initLogger)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
}

// Initialize configuration
func initConfig() {
	fmt.Println("Initializing configuration: initConfig")

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

	// Initialize configuration
	var configErr error
	globalConfig, configErr = config.NewConfig() // This should be the only place you call NewConfig
	if configErr != nil {
		fmt.Println("Error creating Config", "error", configErr)
	}
}

func initLogger() {
	env := globalConfig.App.Environment
	if env == "" {
		env = "development" // Set a default environment
	}

	var loggerErr error
	if env == "development" {
		globalLogger, loggerErr = logger.NewDevelopmentLogger() // Use colored logger in development
	} else {
		globalLogger, loggerErr = logger.NewProductionLogger(globalConfig) // Use a different logger for production
	}

	if loggerErr != nil {
		fmt.Fprintln(os.Stderr, "Error creating Logger:", loggerErr)
		os.Exit(1)
	}
}

// Shutdown gracefully shuts down the application
func Shutdown() error {
	// Implement shutdown logic if necessary
	return nil
}
