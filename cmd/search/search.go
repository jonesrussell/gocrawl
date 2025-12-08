// Package search implements the search command for querying content in Elasticsearch.
package search

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
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
	// Logger provides logging capabilities for the search operation
	Logger logger.Interface
	// Config holds the application configuration
	Config config.Interface
	// SearchManager is the service responsible for executing searches
	SearchManager api.SearchManager
	// IndexName specifies which Elasticsearch index to search
	IndexName string
	// Query contains the search query string
	Query string
	// ResultSize determines how many results to return
	ResultSize int
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
func runSearch(cmd *cobra.Command, _ []string) error {
	// Get dependencies
	deps, err := common.NewCommandDeps()
	if err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Convert size string to int
	sizeStr := cmd.Flag("size").Value.String()
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidSize, err)
	}

	// Get command-line parameters
	indexName := cmd.Flag("index").Value.String()
	queryStr := cmd.Flag("query").Value.String()

	// Create storage using common function
	storageResult, err := common.CreateStorage(deps.Config, deps.Logger)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create search manager
	searchManager := storage.NewSearchManager(storageResult.Storage, deps.Logger)

	// Create params and execute search
	params := Params{
		Logger:        deps.Logger,
		Config:        deps.Config,
		SearchManager: searchManager,
		IndexName:     indexName,
		Query:         queryStr,
		ResultSize:    size,
	}

	// Execute search
	return ExecuteSearch(cmd.Context(), params)
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
func processSearchResults(rawResults []any, log logger.Interface) []Result {
	var results []Result
	for _, raw := range rawResults {
		hit, ok := raw.(map[string]any)
		if !ok {
			log.Error("Failed to convert search result to map",
				"error", "type assertion failed",
				"got_type", fmt.Sprintf("%T", raw),
				"expected_type", "map[string]any")
			continue
		}

		source, ok := hit["_source"].(map[string]any)
		if !ok {
			log.Error("Failed to extract _source from hit",
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

	const (
		indexColumnNumber   = 1
		indexColumnWidth    = 4
		urlColumnNumber     = 2
		contentColumnNumber = 3
		urlColumnWidthRatio = 3
		contentColumnRatio  = 3
	)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: indexColumnNumber, WidthMax: indexColumnWidth}, // Index column (#)
		// URL column (1/3 of table width)
		{Number: urlColumnNumber, WidthMax: DefaultTableWidth / urlColumnWidthRatio},
		// Content preview column (2/3 of table width)
		{Number: contentColumnNumber, WidthMax: DefaultTableWidth * 2 / contentColumnRatio},
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
