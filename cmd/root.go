// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crawlcmd "github.com/jonesrussell/gocrawl/cmd/crawl"
	httpdcmd "github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/indices"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/cmd/sources"
)

var (
	// cfgFile holds the path to the configuration file.
	// It can be set via the --config flag or defaults to config.yaml.
	cfgFile string

	// debug enables debug mode for all commands
	debug bool

	// rootCmd represents the root command for the GoCrawl CLI.
	// It serves as the base command that all subcommands are attached to.
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler for collecting and processing content",
		Long: `gocrawl is a web crawler that helps you collect and process content from various sources.
It provides a flexible and extensible framework for building custom crawlers.`,
		PersistentPreRunE: setupConfig,
	}
)

// setupConfig handles configuration file setup for all commands.
// It ensures the config file path is absolute and configures Viper.
func setupConfig(cmd *cobra.Command, args []string) error {
	// Get the absolute path to the config file
	if cfgFile != "" {
		absPath, err := filepath.Abs(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for config file: %w", err)
		}
		cfgFile = absPath
	} else {
		// Default to config.yaml in the current directory
		cfgFile = "config.yaml"
	}

	// Initialize Viper
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Set the command name in the configuration
	commandName := cmd.Name()
	if cmd.Parent() != nil {
		commandName = fmt.Sprintf("%s %s", cmd.Parent().Name(), cmd.Name())
	}
	viper.Set("command", commandName)

	// Print debug information
	fmt.Printf("Config file: %s\n", cfgFile)
	fmt.Printf("Config file exists: %v\n", viper.ConfigFileUsed() != "")
	fmt.Printf("Crawler config: %+v\n", viper.Get("crawler"))

	return nil
}

// Execute is the entry point for the CLI application.
// It runs the root command and handles any errors that occur during execution.
// If an error occurs, it prints the error message and exits with status code 1.
func Execute() {
	log, err := logger.New(logger.DefaultConfig())
	if err != nil {
		os.Exit(1)
	}

	// Add commands
	rootCmd.AddCommand(
		job.NewJobCommand(log),         // Main job command with fx
		job.NewJobSubCommands(log),     // Job subcommands
		indices.Command(),              // For managing Elasticsearch indices
		sources.NewSourcesCommand(log), // For managing web content sources
		crawlcmd.Command(),             // For crawling web content
		httpdcmd.Command(),             // For running the HTTP server
		search.Command(),               // For searching content in Elasticsearch
	)

	if executeErr := rootCmd.Execute(); executeErr != nil {
		log.Error("Failed to execute command", "error", executeErr)
		os.Exit(1)
	}
}

// init initializes the root command and its subcommands.
// It sets up:
// - The persistent --config flag for specifying the configuration file
// - The persistent --debug flag for enabling debug mode
// - Adds all subcommands for managing different aspects of the crawler:
//   - indices: For managing Elasticsearch indices
//   - sources: For managing web content sources
//   - crawl: For crawling web content
//   - httpd: For running the HTTP server
//   - job: For managing scheduled crawl jobs
//   - search: For searching content in Elasticsearch
func init() {
	// Add the persistent --config flag to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is config.yaml)")

	// Add the persistent --debug flag to all commands
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
}
