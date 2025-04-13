// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
	DefaultTableWidth = 160

	// Table column configuration constants
	columnNumberIndex       = 1
	columnNumberURL         = 2
	columnNumberContent     = 3
	columnWidthIndex        = 4
	columnWidthURLRatio     = 3 // URL column takes 1/3 of table width
	columnWidthContentRatio = 3 // Content column takes 2/3 of table width
)

// Error constants
const (
	ErrLoggerNotFound = "logger not found in context or invalid type"
	ErrConfigNotFound = "config not found in context or invalid type"
	ErrInvalidSize    = "invalid size value"
	ErrStartFailed    = "failed to start application"
	ErrStopFailed     = "failed to stop application"
)

// Params holds the search operation parameters
type Params struct {
	fx.In

	// Logger provides logging capabilities for the search operation
	Logger logger.Interface
	// Config holds the application configuration
	Config config.Interface
	// SearchManager is the service responsible for executing searches
	SearchManager api.SearchManager
	// IndexName specifies which Elasticsearch index to search
	IndexName string `name:"searchIndex"`
	// Query contains the search query string
	Query string `name:"searchQuery"`
	// ResultSize determines how many results to return
	ResultSize int `name:"searchSize"`
}

// Result represents a search result
type Result struct {
	URL     string
	Content string
}

// Cmd represents the search command that allows users to search content
// in Elasticsearch using various parameters.
var Cmd = &cobra.Command{
	Use:   "search",
	Short: "Search content in Elasticsearch",
	Long: `Search command allows you to search through crawled content in Elasticsearch.

Examples:
  # Search for content containing "golang"
  gocrawl search -q "golang"

  # Search in a specific index with custom result size
  gocrawl search -i "articles" -q "golang" -s 20

Flags:
  -i, --index string   Index to search (default "articles")
  -q, --query string   Query string to search for (required)
  -s, --size int      Number of results to return (default 10)
`,
	RunE: runSearch,
}

// searchModule provides the search command dependencies
var searchModule = Module

// Command returns the search command for use in the root command
func Command() *cobra.Command {
	// Define flags for the search command
	Cmd.Flags().StringP("index", "i", "articles", "Index to search")
	Cmd.Flags().IntP("size", "s", DefaultSearchSize, "Number of results to return")
	Cmd.Flags().StringP("query", "q", "", "Query string to search for")

	// Mark the query flag as required
	if err := Cmd.MarkFlagRequired("query"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking query flag as required: %v\n", err)
		os.Exit(1)
	}

	return Cmd
}

// runSearch executes the search command with the provided parameters.
// It handles:
// - Parameter validation and retrieval
// - Signal handling for graceful shutdown
// - Application lifecycle management
// - Search execution and result display
func runSearch(cmd *cobra.Command, _ []string) error {
	// Get logger from context
	loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
	log, ok := loggerValue.(logger.Interface)
	if !ok {
		return errors.New(ErrLoggerNotFound)
	}

	// Convert size string to int
	sizeStr := cmd.Flag("size").Value.String()
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidSize, err)
	}

	// Create Fx app with the module
	fxApp := fx.New(
		searchModule,
		fx.Provide(
			func() logger.Interface { return log },
			fx.Annotate(
				func() string { return cmd.Flag("index").Value.String() },
				fx.ResultTags(`name:"searchIndex"`),
			),
			fx.Annotate(
				func() string { return cmd.Flag("query").Value.String() },
				fx.ResultTags(`name:"searchQuery"`),
			),
			fx.Annotate(
				func() int { return size },
				fx.ResultTags(`name:"searchSize"`),
			),
		),
		fx.Invoke(func(p Params) error {
			return ExecuteSearch(cmd.Context(), p)
		}),
		fx.WithLogger(func() fxevent.Logger {
			return logger.NewFxLogger(log)
		}),
	)

	// Start the application
	log.Info("Starting search")
	startErr := fxApp.Start(cmd.Context())
	if startErr != nil {
		log.Error("Failed to start application", "error", startErr)
		return fmt.Errorf("%s: %w", ErrStartFailed, startErr)
	}

	// Wait for interrupt signal
	log.Info("Waiting for interrupt signal")
	<-cmd.Context().Done()

	// Stop the application
	log.Info("Stopping application")
	stopErr := fxApp.Stop(cmd.Context())
	if stopErr != nil {
		log.Error("Failed to stop application", "error", stopErr)
		return fmt.Errorf("%s: %w", ErrStopFailed, stopErr)
	}

	log.Info("Application stopped successfully")
	return nil
}

// buildSearchQuery constructs the Elasticsearch query
func buildSearchQuery(size int, query string) map[string]any {
	return map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": query,
			},
		},
		"size": size,
	}
}

// processSearchResults converts raw search results to Result structs
func processSearchResults(rawResults []any, logger logger.Interface) []Result {
	var results []Result
	for _, raw := range rawResults {
		hit, ok := raw.(map[string]any)
		if !ok {
			logger.Error("Failed to convert search result to map",
				"error", "type assertion failed",
				"got_type", fmt.Sprintf("%T", raw),
				"expected_type", "map[string]any")
			continue
		}

		source, ok := hit["_source"].(map[string]any)
		if !ok {
			logger.Error("Failed to extract _source from hit",
				"error", "type assertion failed",
				"got_type", fmt.Sprintf("%T", hit["_source"]),
				"expected_type", "map[string]any")
			continue
		}

		url, _ := source["url"].(string)
		content, _ := source["content"].(string)

		results = append(results, Result{
			URL:     url,
			Content: content,
		})
	}
	return results
}

// configureResultsTable sets up the table writer with appropriate styling and columns
func configureResultsTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateRows = true

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: columnNumberIndex, WidthMax: columnWidthIndex},
		{Number: columnNumberURL, WidthMax: DefaultTableWidth / columnWidthURLRatio},
		{Number: columnNumberContent, WidthMax: DefaultTableWidth * 2 / columnWidthContentRatio},
	})

	t.AppendHeader(table.Row{"#", "URL", "Content Preview"})
	return t
}

// renderSearchResults formats and displays the search results in a table
func renderSearchResults(results []Result, query string) {
	t := configureResultsTable()

	for i, result := range results {
		content := strings.TrimSpace(result.Content)
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.Join(strings.Fields(content), " ")
		contentPreview := truncateString(content, DefaultContentPreviewLength)

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

	t.AppendFooter(table.Row{"Total", len(results), fmt.Sprintf("Query: %s", query)})

	fmt.Fprintf(os.Stdout, "\nSearch Results:\n")
	t.Render()
}

// ExecuteSearch performs the actual search operation using the provided parameters.
func ExecuteSearch(ctx context.Context, p Params) error {
	p.Logger.Info("Starting search...",
		"query", p.Query,
		"index", p.IndexName,
		"size", p.ResultSize,
	)

	query := buildSearchQuery(p.ResultSize, p.Query)
	rawResults, err := p.SearchManager.Search(ctx, p.IndexName, query)
	if err != nil {
		p.Logger.Error("Search failed", "error", err)
		return fmt.Errorf("search failed: %w", err)
	}

	results := processSearchResults(rawResults, p.Logger)

	p.Logger.Info("Search completed",
		"query", p.Query,
		"results", len(results),
	)

	if len(results) == 0 {
		fmt.Fprintf(os.Stdout, "No results found for query: %s\n", p.Query)
	}

	// Close the search manager
	if closeErr := p.SearchManager.Close(); closeErr != nil {
		p.Logger.Error("Error closing search manager", "error", closeErr)
		return fmt.Errorf("error closing search manager: %w", closeErr)
	}

	renderSearchResults(results, p.Query)
	return nil
}

// truncateString truncates a string to the specified length and adds ellipsis if needed
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
