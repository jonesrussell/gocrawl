package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	DefaultSearchSize = 10 // Default number of results to return
)

// SearchParams holds the parameters for the search operation
type SearchParams struct {
	fx.In

	Logger     common.Logger
	Config     common.Config
	SearchSvc  *search.Service
	IndexName  string `name:"indexName"`
	Query      string `name:"query"`
	ResultSize int    `name:"resultSize"`
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search content in Elasticsearch",
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, _ []string) error {
	queryStr, queryErr := cmd.Flags().GetString("query")
	if queryErr != nil {
		return fmt.Errorf("error retrieving query: %w", queryErr)
	}

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

	app := fx.New(
		common.Module,
		search.Module,
		fx.Provide(
			// Provide search parameters
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

func executeSearch(ctx context.Context, p SearchParams) error {
	p.Logger.Info("Starting search...",
		"query", p.Query,
		"index", p.IndexName,
		"size", p.ResultSize,
	)

	results, err := p.SearchSvc.SearchContent(ctx, p.Query, p.IndexName, p.ResultSize)
	if err != nil {
		p.Logger.Error("Search failed", "error", err)
		return fmt.Errorf("search failed: %w", err)
	}

	// Print results
	if len(results) == 0 {
		common.PrintInfof("No results found")
		return nil
	}

	common.PrintInfof("\nFound %d results:", len(results))
	for i, result := range results {
		common.PrintInfof("\nResult %d:", i+1)
		common.PrintInfof("URL: %s", result.URL)
		common.PrintInfof("Content: %s", result.Content)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Define flags for the search command
	searchCmd.Flags().StringP("index", "i", "articles", "Index to search")
	searchCmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	searchCmd.Flags().StringP("query", "q", "", "Query string to search for")

	if err := searchCmd.MarkFlagRequired("query"); err != nil {
		common.PrintErrorf("Error marking query flag as required: %v", err)
		os.Exit(1)
	}
}
