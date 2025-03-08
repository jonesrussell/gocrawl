// Package cmd implements the command-line interface for GoCrawl.
// This file contains the search command implementation that allows users to search
// content in Elasticsearch using various parameters.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	// DefaultSearchSize defines the default number of search results to return
	// when no size is specified via command-line flags
	DefaultSearchSize = 10
)

// SearchParams holds the parameters required for executing a search operation.
// It uses fx.In for dependency injection of required components.
type SearchParams struct {
	fx.In

	// Logger provides logging capabilities for the search operation
	Logger logger.Interface
	// Config holds the application configuration
	Config common.Config
	// SearchSvc is the service responsible for executing searches
	SearchSvc *search.Service
	// IndexName specifies which Elasticsearch index to search
	IndexName string `name:"indexName"`
	// Query contains the search query string
	Query string `name:"query"`
	// ResultSize determines how many results to return
	ResultSize int `name:"resultSize"`
}

// searchCmd represents the search command that allows users to search content
// in Elasticsearch using various parameters.
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search content in Elasticsearch",
	RunE:  runSearch,
}

// runSearch executes the search command with the provided parameters.
// It handles:
// - Parameter validation and retrieval
// - Signal handling for graceful shutdown
// - Application lifecycle management
// - Search execution and result display
func runSearch(cmd *cobra.Command, _ []string) error {
	// Retrieve and validate the search query
	queryStr, queryErr := cmd.Flags().GetString("query")
	if queryErr != nil {
		return fmt.Errorf("error retrieving query: %w", queryErr)
	}

	// Get the index name and result size from flags
	indexName := cmd.Flag("index").Value.String()
	size, sizeErr := cmd.Flags().GetInt("size")
	if sizeErr != nil {
		size = DefaultSearchSize
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to signal search completion
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules and dependencies
	app := fx.New(
		common.Module,
		search.Module,
		fx.Provide(
			// Provide search parameters with appropriate tags for dependency injection
			fx.Annotate(
				func() string { return queryStr },
				fx.ResultTags(`name:"query"`),
			),
			fx.Annotate(
				func() string { return indexName },
				fx.ResultTags(`name:"indexName"`),
			),
			fx.Annotate(
				func() int { return size },
				fx.ResultTags(`name:"resultSize"`),
			),
		),
		fx.Invoke(func(p SearchParams) {
			// Execute the search and handle any errors
			if startErr := executeSearch(cmd.Context(), p); startErr != nil {
				p.Logger.Error("Error executing search", "error", startErr)
			}
			close(doneChan) // Signal completion
		}),
	)

	// Start the application and handle any startup errors
	if err := app.Start(cmd.Context()); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Wait for either:
	// - A signal interrupt (SIGINT/SIGTERM)
	// - Context cancellation
	// - Search completion
	select {
	case sig := <-sigChan:
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
	case <-cmd.Context().Done():
		common.PrintInfof("\nContext cancelled, initiating shutdown...")
	case <-doneChan:
		common.PrintInfof("\nSearch completed, shutting down...")
	}

	// Create a context with timeout for graceful shutdown
	stopCtx, stopCancel := context.WithTimeout(cmd.Context(), common.DefaultOperationTimeout)
	defer stopCancel()

	// Stop the application and handle any shutdown errors
	if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		common.PrintErrorf("Error stopping application: %v", err)
		return err
	}

	return nil
}

// executeSearch performs the actual search operation using the provided parameters.
// It:
// - Logs the start of the search operation
// - Executes the search using the search service
// - Handles and logs any errors
// - Displays the search results
func executeSearch(ctx context.Context, p SearchParams) error {
	// Log the start of the search operation with parameters
	p.Logger.Info("Starting search...",
		"query", p.Query,
		"index", p.IndexName,
		"size", p.ResultSize,
	)

	// Execute the search and handle any errors
	results, err := p.SearchSvc.SearchContent(ctx, p.Query, p.IndexName, p.ResultSize)
	if err != nil {
		p.Logger.Error("Search failed", "error", err)
		return fmt.Errorf("search failed: %w", err)
	}

	// Handle empty results
	if len(results) == 0 {
		common.PrintInfof("No results found")
		return nil
	}

	// Display search results
	common.PrintInfof("\nFound %d results:", len(results))
	for i, result := range results {
		common.PrintInfof("\nResult %d:", i+1)
		common.PrintInfof("URL: %s", result.URL)
		common.PrintInfof("Content: %s", result.Content)
	}

	return nil
}

// init initializes the search command by:
// - Adding it to the root command
// - Setting up command-line flags
// - Marking required flags
func init() {
	rootCmd.AddCommand(searchCmd)

	// Define flags for the search command
	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	// Mark the query flag as required
	if err := searchCmd.MarkFlagRequired("query"); err != nil {
		common.PrintErrorf("Error marking query flag as required: %v", err)
		os.Exit(1)
	}
}
