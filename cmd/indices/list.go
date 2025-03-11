// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. This file contains the implementation of the list command
// that displays all indices in a formatted table with their health status and metrics.
package indices

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
// It uses fx.In to enable dependency injection of required components.
// This struct is used internally by the executeList function to manage
// dependencies and maintain clean separation of concerns.
type listParams struct {
	fx.In

	// ctx is the context for managing timeouts and cancellation
	ctx context.Context

	// storage provides access to Elasticsearch operations
	storage common.Storage

	// logger is used for structured logging throughout the command
	logger common.Logger
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
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules
	app := fx.New(
		common.Module,
		Module,
		fx.Invoke(func(lc fx.Lifecycle, s common.Storage, l common.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &listParams{
						ctx:     cmd.Context(),
						storage: s,
						logger:  l,
					}
					if err := executeList(params); err != nil {
						l.Error("Error executing list", "error", err)
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
	// - List completion
	// - List error
	var listErr error
	select {
	case sig := <-sigChan:
		common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
	case <-cmd.Context().Done():
		common.PrintInfof("\nContext cancelled, initiating shutdown...")
	case listErr = <-errChan:
		// Error already printed in executeList
	case <-doneChan:
		// Success message already printed in executeList
	}

	// Create a context with timeout for graceful shutdown
	stopCtx, stopCancel := context.WithTimeout(cmd.Context(), common.DefaultOperationTimeout)
	defer stopCancel()

	// Stop the application and handle any shutdown errors
	if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
		common.PrintErrorf("Error stopping application: %v", err)
		if listErr == nil {
			listErr = err
		}
	}

	return listErr
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
		t.AppendRow([]interface{}{
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
