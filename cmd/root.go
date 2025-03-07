package cmd

import (
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/cmd/indices"
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

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")
}

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
	var loggerErr error
	globalLogger, loggerErr = logger.InitializeLogger(globalConfig)
	if loggerErr != nil {
		fmt.Fprintf(os.Stderr, "Error creating Logger: %v\n", loggerErr)
		os.Exit(1)
	}
	shared.SetLogger(globalLogger)
}

func init() {
	rootCmd.AddCommand(indices.Command())
	rootCmd.AddCommand(sources.Command())
}
