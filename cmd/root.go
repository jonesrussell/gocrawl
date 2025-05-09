// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jonesrussell/gocrawl/cmd/common"
	crawlcmd "github.com/jonesrussell/gocrawl/cmd/crawl"
	httpdcmd "github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/index"
	"github.com/jonesrussell/gocrawl/cmd/scheduler"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/config"
)

var (
	// cfgFile holds the path to the configuration file.
	// It can be set via the --config flag or defaults to config.yaml.
	cfgFile string

	// Debug enables debug mode for all commands
	Debug bool

	// rootCmd represents the root command for the GoCrawl CLI.
	// It serves as the base command that all subcommands are attached to.
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler for collecting and processing content",
		Long: `gocrawl is a web crawler that helps you collect and process content from various sources.
It provides a flexible and extensible framework for building custom crawlers.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logger
			log, err := logger.NewFromConfig(viper.GetViper())
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			// Create a context with dependencies
			ctx := context.Background()
			ctx = context.WithValue(ctx, common.LoggerKey, log)
			ctx = context.WithValue(ctx, common.ConfigKey, cfg)
			cmd.SetContext(ctx)

			return nil
		},
	}
)

// bindFlags binds all command flags to viper configuration.
func bindFlags(cmd *cobra.Command) error {
	// Bind persistent flags
	if err := viper.BindPFlag("app.debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		return fmt.Errorf("failed to bind debug flag: %w", err)
	}

	if err := viper.BindPFlag("app.config_file", cmd.PersistentFlags().Lookup("config")); err != nil {
		return fmt.Errorf("failed to bind config flag: %w", err)
	}

	// Bind command-specific flags
	if cmd.Name() == "crawl" {
		if err := viper.BindPFlag("crawler.max_depth", cmd.Flags().Lookup("depth")); err != nil {
			return fmt.Errorf("failed to bind depth flag: %w", err)
		}
		if err := viper.BindPFlag("crawler.rate_limit", cmd.Flags().Lookup("rate-limit")); err != nil {
			return fmt.Errorf("failed to bind rate-limit flag: %w", err)
		}
	}

	return nil
}

// Execute is the entry point for the CLI application.
// It runs the root command and handles any errors that occur during execution.
// If an error occurs, it prints the error message and exits with status code 1.
func Execute() {
	// Add commands first
	rootCmd.AddCommand(
		scheduler.Command(),         // Main scheduler command with fx
		index.Command(),             // For managing Elasticsearch index
		sources.NewSourcesCommand(), // For managing web content sources
		crawlcmd.Command(),          // For crawling web content
		httpdcmd.Command(),          // For running the HTTP server
		search.Command(),            // For searching content in Elasticsearch
	)

	if executeErr := rootCmd.Execute(); executeErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", executeErr)
		os.Exit(1)
	}
}

// init initializes the root command and its subcommands.
func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is ./config.yaml, ~/.crawler/config.yaml, or /etc/crawler/config.yaml)",
	)
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug mode")

	// Bind flags
	if err := bindFlags(rootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flags: %v\n", err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in default locations
		viper.AddConfigPath(".")              // Current directory
		viper.AddConfigPath("$HOME/.gocrawl") // User config directory
		viper.AddConfigPath("/etc/gocrawl")   // System config directory
		viper.SetConfigName("config")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()
}
