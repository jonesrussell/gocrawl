// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"os"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for table formatting
const (
	// HeaderWidth defines the width of the header divider line
	HeaderWidth = 17
	// TableWidth defines the width of the table divider line
	TableWidth = 92
)

// listParams holds the parameters required for displaying the sources list.
// It contains the context, sources configuration, and logger needed for
// the list operation.
type listParams struct {
	// ctx is the context for the list operation
	ctx context.Context
	// sources contains the configuration for all content sources
	sources *sources.Sources
	// logger provides logging capabilities for the list operation
	logger common.Logger
}

// listCommand creates and returns the list command that displays all configured sources.
// It:
// - Sets up the command with appropriate usage and description
// - Configures the command to use runList as its execution function
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long:  `Display a list of all sources configured in sources.yml.`,
		Run:   runList,
	}
}

// runList executes the list command and displays all configured sources.
// It:
// - Initializes the Fx application with required modules
// - Sets up context with timeout for graceful shutdown
// - Handles application lifecycle and error cases
// - Displays the sources list in a formatted table
func runList(cmd *cobra.Command, _ []string) {
	var logger common.Logger
	var exitCode int

	// Initialize the Fx application with required modules
	app := fx.New(
		common.Module,
		fx.Invoke(func(s *sources.Sources, l common.Logger) {
			logger = l
			params := &listParams{
				ctx:     cmd.Context(),
				sources: s,
				logger:  l,
			}
			displaySourcesList(params)
		}),
	)

	// Set up context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(cmd.Context(), common.DefaultStartupTimeout)
	defer func() {
		cancel()
		if err := app.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
			if logger != nil {
				logger.Error("Error stopping application", "error", err)
				exitCode = 1
			}
		}
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

// displaySourcesList formats and displays the list of configured sources.
// It:
// - Prints a header with the command title
// - Displays a table header with column names
// - Lists each source with its configuration details
// - Uses consistent formatting for better readability
func displaySourcesList(p *listParams) {
	common.PrintInfof("\nConfigured Sources")
	common.PrintDivider(HeaderWidth)
	common.PrintTableHeaderf("%-20s %-30s %-15s %-15s %-10s",
		"Name", "URL", "Article Index", "Content Index", "Max Depth")
	common.PrintDivider(TableWidth)

	// Display each source in a formatted table row
	for _, source := range p.sources.Sources {
		common.PrintTableHeaderf("%-20s %-30s %-15s %-15s %-10d",
			source.Name,
			source.URL,
			source.ArticleIndex,
			source.Index,
			source.MaxDepth)
	}

	common.PrintInfof("")
}
