// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for table formatting
const (
	// TableWidth defines the total width of the table output for consistent formatting
	TableWidth = 92
)

// listParams holds the parameters required for listing indices.
// It contains the context, storage interface, and logger needed for
// the list operation.
type listParams struct {
	fx.In

	ctx     context.Context
	storage common.Storage
	logger  common.Logger
}

// listCommand creates and returns the list command that displays all indices.
// It:
// - Sets up the command with appropriate usage and description
// - Configures the command to use runList as its execution function
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		Run: runList,
	}
}

// runList executes the list command and displays all indices.
// It:
// - Initializes the Fx application with required modules
// - Sets up context with timeout for graceful shutdown
// - Handles application lifecycle and error cases
// - Displays the indices list in a formatted table
func runList(cmd *cobra.Command, _ []string) {
	var logger common.Logger
	var exitCode int

	// Initialize the Fx application with required modules
	app := fx.New(
		common.Module,
		fx.Invoke(func(s common.Storage, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				storage: s,
				logger:  l,
			}
			if err := executeList(params); err != nil {
				l.Error("Error executing list", "error", err)
				exitCode = 1
			}
		}),
	)

	// Set up context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer func() {
		if err := app.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			if logger != nil {
				logger.Error("Error stopping application", "error", err)
				exitCode = 1
			}
		}
		cancel()
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	// Start the application and handle any startup errors
	if err := app.Start(ctx); err != nil {
		if logger != nil {
			logger.Error("Error starting application", "error", err)
		}
		exitCode = 1
		return
	}
}

// executeList retrieves and displays the list of indices.
// It:
// - Retrieves all indices from Elasticsearch
// - Filters out internal indices (starting with '.')
// - Handles empty results
// - Displays the indices in a formatted table
func executeList(p *listParams) error {
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return err
	}

	// Filter out internal indices
	var filteredIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	if len(filteredIndices) == 0 {
		p.logger.Info("No indices found")
		return nil
	}

	return printIndices(p.ctx, filteredIndices, p.storage, p.logger)
}

// printIndices formats and displays the indices in a table.
// It:
// - Creates a new table with appropriate headers
// - Retrieves health status and document count for each index
// - Handles errors gracefully
// - Renders the table with all index information
func printIndices(ctx context.Context, indices []string, storage common.Storage, logger common.Logger) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Ingestion Name", "Ingestion Status"})

	for _, index := range indices {
		healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			logger.Error("Error getting health for index", "index", index, "error", healthErr)
			continue
		}

		docCount, docErr := storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			docCount = 0
		}

		ingestionStatus := getIngestionStatus(healthStatus)

		t.AppendRow([]interface{}{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for ingestion name (not implemented yet)
			ingestionStatus,
		})
	}

	if t.Length() == 0 {
		logger.Info("No indices found")
		return nil
	}

	t.Render()
	return nil
}

// getIngestionStatus maps the index health status to a human-readable
// ingestion status. It:
// - Maps "red" to "Disconnected"
// - Maps "yellow" to "Warning"
// - Maps other statuses to "Connected"
func getIngestionStatus(healthStatus string) string {
	switch healthStatus {
	case "red":
		return "Disconnected"
	case "yellow":
		return "Warning"
	default:
		return "Connected"
	}
}
