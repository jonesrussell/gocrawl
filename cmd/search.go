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
	"strings"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/search"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for default values
const (
	// DefaultSearchSize defines the default number of search results to return
	// when no size is specified via command-line flags
	DefaultSearchSize = 10

	// DefaultContentPreviewLength defines the maximum length for content previews
	// in search results before truncation
	DefaultContentPreviewLength = 100

	// DefaultTableWidth defines the maximum width for the content preview column
	DefaultTableWidth = 80

	// Table column configuration constants
	columnNumberIndex       = 1
	columnNumberURL         = 2
	columnNumberContent     = 3
	columnWidthIndex        = 4
	columnWidthURLRatio     = 3 // URL column takes 1/3 of table width
	columnWidthContentRatio = 3 // Content column takes 2/3 of table width
)

// SearchParams holds the parameters required for executing a search operation.
// It uses fx.In for dependency injection of required components.
type SearchParams struct {
	fx.In

	// Logger provides logging capabilities for the search operation
	Logger common.Logger
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

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
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
		fx.Invoke(func(lc fx.Lifecycle, p SearchParams) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					// Execute the search and handle any errors
					if err := executeSearch(cmd.Context(), p); err != nil {
						p.Logger.Error("Error executing search", "error", err)
						common.PrintErrorf("\nSearch failed: %v", err)
						errChan <- err
						return err
					}
					close(doneChan)
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
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
	// - Search error
	var searchErr error
	select {
	case sig := <-sigChan:
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
	case <-cmd.Context().Done():
		common.PrintInfof("\nContext cancelled, initiating shutdown...")
	case searchErr = <-errChan:
		// Error already printed in executeSearch
	case <-doneChan:
		// Success message already printed in executeSearch
	}

	// Create a context with timeout for graceful shutdown
	stopCtx, stopCancel := context.WithTimeout(cmd.Context(), common.DefaultOperationTimeout)
	defer stopCancel()

	// Stop the application and handle any shutdown errors
	if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		common.PrintErrorf("Error stopping application: %v", err)
		return err
	}

	return searchErr
}

// executeSearch performs the actual search operation using the provided parameters.
// It:
// - Logs the start of the search operation
// - Executes the search using the search service
// - Handles and logs any errors
// - Displays the search results in a formatted table
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

	// Log completion with result count
	p.Logger.Info("Search completed",
		"query", p.Query,
		"results", len(results),
	)

	// Handle empty results
	if len(results) == 0 {
		common.PrintInfof("No results found for query: %s", p.Query)
		return nil
	}

	// Create and configure the table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateRows = true

	// Configure column widths
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: columnNumberIndex, WidthMax: columnWidthIndex},
		{Number: columnNumberURL, WidthMax: DefaultTableWidth / columnWidthURLRatio},
		{Number: columnNumberContent, WidthMax: DefaultTableWidth * 2 / columnWidthContentRatio},
	})

	// Set up the headers
	t.AppendHeader(table.Row{"#", "URL", "Content Preview"})

	// Add each result as a row
	for i, result := range results {
		// Clean and format the content preview
		content := strings.TrimSpace(result.Content)
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.Join(strings.Fields(content), " ")
		contentPreview := truncateString(content, DefaultContentPreviewLength)

		// Clean and format the URL
		url := strings.TrimSpace(result.URL)
		if url == "" {
			url = "N/A"
		}

		t.AppendRow(table.Row{
			i + 1,
			url,
			contentPreview,
		})
	}

	// Add a footer with summary information
	t.AppendFooter(table.Row{"Total", len(results), fmt.Sprintf("Query: %s", p.Query)})

	// Print a header
	common.PrintInfof("\nSearch Results:")
	// Render the table
	t.Render()

	// Print success message
	common.PrintInfof("\nSearch completed successfully with %d results.", len(results))

	return nil
}

// truncateString truncates a string to the specified length and adds ellipsis if needed
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
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
