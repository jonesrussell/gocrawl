// Package cmd implements the command-line interface for GoCrawl.
// It provides the root command and subcommands for managing web crawling operations.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	crawlcmd "github.com/jonesrussell/gocrawl/cmd/crawl"
	httpdcmd "github.com/jonesrussell/gocrawl/cmd/httpd"
	"github.com/jonesrussell/gocrawl/cmd/indices"
	jobcmd "github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/cmd/search"
	"github.com/jonesrussell/gocrawl/cmd/sources"
	"github.com/jonesrussell/gocrawl/internal/common"
)

var (
	// cfgFile holds the path to the configuration file.
	// It can be set via the --config flag or defaults to config.yaml.
	cfgFile string

	// rootCmd represents the root command for the GoCrawl CLI.
	// It serves as the base command that all subcommands are attached to.
	rootCmd = &cobra.Command{
		Use:   "gocrawl",
		Short: "A web crawler that stores content in Elasticsearch",
		Long: `GoCrawl is a powerful web crawler that efficiently collects and stores web content
in Elasticsearch. It supports configurable crawling strategies, rate limiting,
and content processing through a YAML-based configuration system.

The crawler can be configured to:
- Define multiple content sources with custom crawling rules
- Process and extract structured data from web pages
- Store content in Elasticsearch with proper indexing
- Handle rate limiting and respect robots.txt
- Process different types of content (articles, general web pages)`,
		PersistentPreRunE: setupConfig,
	}
)

// setupConfig handles configuration file setup for all commands.
// It ensures the config file path is absolute and sets it in the environment.
func setupConfig(_ *cobra.Command, _ []string) error {
	// If config file is provided via flag, use absolute path
	if cfgFile != "" {
		if !os.IsPathSeparator(cfgFile[0]) {
			// Convert relative path to absolute
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			cfgFile = wd + string(os.PathSeparator) + cfgFile
		}
		err := os.Setenv("CONFIG_FILE", cfgFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute is the entry point for the CLI application.
// It runs the root command and handles any errors that occur during execution.
// If an error occurs, it prints the error message and exits with status code 1.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		common.PrintErrorf("Error executing root command: %v", err)
		os.Exit(1)
	}
}

// init initializes the root command and its subcommands.
// It sets up:
// - The persistent --config flag for specifying the configuration file
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

	// Add subcommands for managing different aspects of the crawler
	rootCmd.AddCommand(
		indices.Command(),  // For managing Elasticsearch indices
		sources.Command(),  // For managing web content sources
		crawlcmd.Command(), // For crawling web content
		httpdcmd.Command(), // For running the HTTP server
		jobcmd.Command(),   // For managing scheduled crawl jobs
		search.Command(),   // For searching content in Elasticsearch
	)
}
