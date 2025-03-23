// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for table formatting
const (
	// TableWidth defines the total width of the table output for consistent formatting.
	// This ensures that the output remains readable across different terminal sizes.
	TableWidth = 92
)

// listParams holds the parameters required for listing indices.
type listParams struct {
	ctx     context.Context
	storage common.Storage
	logger  common.Logger
}

// listCommand creates and returns the list command that displays all indices.
// It integrates with the Cobra command framework and sets up the command
// structure with appropriate usage information and examples.
//
// Returns:
//   - *cobra.Command: A configured Cobra command ready to be added to the root command
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all Elasticsearch indices",
		Long: `Display a list of all indices in the Elasticsearch cluster.

Example:
  gocrawl indices list`,
		RunE: runList,
	}
}

// runList executes the list command and displays all indices.
// It:
// - Sets up signal handling for graceful shutdown
// - Creates channels for error handling and completion
// - Initializes the Fx application with required modules
// - Handles application lifecycle and error cases
// - Displays the indices list in a formatted table
func runList(cmd *cobra.Command, _ []string) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Set up signal handling
	handler := signal.NewSignalHandler()
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Initialize the Fx application with required modules
	fxApp := fx.New(
		fx.NopLogger, // Suppress Fx startup/shutdown logs
		Module,
		fx.Invoke(func(p struct {
			fx.In
			Storage common.Storage
			Logger  common.Logger
			LC      fx.Lifecycle
		}) {
			p.LC.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &listParams{
						ctx:     ctx,
						storage: p.Storage,
						logger:  p.Logger,
					}
					if err := executeList(params); err != nil {
						p.Logger.Error("Error executing list", "error", err)
						return err
					}
					return nil
				},
				OnStop: func(context.Context) error {
					return nil
				},
			})
		}),
	)

	// Set the fx app for coordinated shutdown
	handler.SetFXApp(fxApp)

	// Start the application
	if err := fxApp.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Wait for shutdown signal
	handler.Wait()

	return nil
}

// executeList retrieves and displays the list of indices.
// This function handles the core business logic of the list command:
// - Retrieving indices from Elasticsearch
// - Filtering internal indices
// - Formatting and displaying results
//
// Parameters:
//   - p: listParams containing all required dependencies
//
// Returns:
//   - error: Any error encountered during execution
func executeList(p *listParams) error {
	// Retrieve all indices from Elasticsearch
	indices, err := p.storage.ListIndices(p.ctx)
	if err != nil {
		return err
	}

	// Filter out internal indices (those starting with '.')
	// This improves readability by showing only relevant indices
	var filteredIndices []string
	for _, index := range indices {
		if !strings.HasPrefix(index, ".") {
			filteredIndices = append(filteredIndices, index)
		}
	}

	// Handle the case where no indices are found
	if len(filteredIndices) == 0 {
		p.logger.Info("No indices found")
		return nil
	}

	return printIndices(p.ctx, filteredIndices, p.storage, p.logger)
}

// printIndices formats and displays the indices in a table format.
// This function handles the presentation layer of the list command:
// - Creates and configures the output table
// - Retrieves additional metadata for each index
// - Handles errors for individual index operations
// - Formats and displays the final output
//
// Parameters:
//   - ctx: Context for managing timeouts and cancellation
//   - indices: List of index names to display
//   - storage: Interface for Elasticsearch operations
//   - logger: Logger for error reporting
//
// Returns:
//   - error: Any error encountered during table rendering
func printIndices(ctx context.Context, indices []string, storage common.Storage, logger common.Logger) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index Name", "Health", "Docs Count", "Ingestion Name", "Ingestion Status"})

	// Process each index and gather its metadata
	for _, index := range indices {
		// Get health status with error handling
		healthStatus, healthErr := storage.GetIndexHealth(ctx, index)
		if healthErr != nil {
			logger.Error("Error getting health for index", "index", index, "error", healthErr)
			continue
		}

		// Get document count with fallback to 0 on error
		docCount, docErr := storage.GetIndexDocCount(ctx, index)
		if docErr != nil {
			docCount = 0
		}

		// Map health status to ingestion status
		ingestionStatus := getIngestionStatus(healthStatus)

		// Add row to table
		t.AppendRow([]any{
			index,
			healthStatus,
			docCount,
			"", // Placeholder for future ingestion name feature
			ingestionStatus,
		})
	}

	// Handle case where no indices could be processed
	if t.Length() == 0 {
		logger.Info("No indices found")
		return nil
	}

	// Render the final table
	t.Render()
	return nil
}

// getIngestionStatus maps Elasticsearch health status to human-readable ingestion status.
// This function provides a user-friendly interpretation of index health:
// - "red" indicates a serious issue (Disconnected)
// - "yellow" indicates a potential issue (Warning)
// - Any other status is considered healthy (Connected)
//
// Parameters:
//   - healthStatus: The Elasticsearch health status string
//
// Returns:
//   - string: A human-readable ingestion status
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
