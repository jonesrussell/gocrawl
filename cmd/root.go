package cmd

import (
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/shared"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	globalLogger logger.Interface
	globalConfig *config.Config
	rootCmd      = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
	}
)

// Execute is the entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		globalLogger.Error("Error executing root command", "error", err)
		os.Exit(1)
	}
}

// Initialize the command
func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")

	// Add commands
	rootCmd.AddCommand(sources.Command())
}

// Initialize configuration
func initConfig() {
	var err error
	globalConfig, err = config.InitializeConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize config: %v\n", err)
		os.Exit(1)
	}
	shared.SetConfig(globalConfig)
}

func initLogger() {
	env := globalConfig.App.Environment
	if env == "" {
		env = "development" // Set a default environment
	}

	// Ensure we have a valid log level
	logLevel := globalConfig.Log.Level
	if logLevel == "" {
		logLevel = "info" // Set a default log level
	}

	// Log the initialization details
	fmt.Fprintf(os.Stderr, "Initializing logger in environment: %s with level: %s\n", env, logLevel)

	var loggerErr error
	if env == "development" {
		globalLogger, loggerErr = logger.NewDevelopmentLogger(logLevel)
	} else {
		globalLogger, loggerErr = logger.NewProductionLogger(logLevel)
	}

	if loggerErr != nil {
		fmt.Fprintf(os.Stderr, "Error creating Logger: %v\n", loggerErr)
		os.Exit(1)
	}

	shared.SetLogger(globalLogger)
}

// Shutdown gracefully shuts down the application
func Shutdown() error {
	// Implement shutdown logic if necessary
	return nil
}
