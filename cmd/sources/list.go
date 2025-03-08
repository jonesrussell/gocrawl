// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// rootCmd represents the sources command
var rootCmd = &cobra.Command{
	Use:   "sources",
	Short: "Manage content sources",
	Long: `Manage content sources in GoCrawl.
This command provides subcommands for listing, adding, and managing content sources.`,
}

// Constants for table formatting
const (
	// TableWidth defines the total width of the table output for consistent formatting
	TableWidth = 92
)

// listParams holds the parameters required for listing sources.
// It contains the context, sources instance, and logger needed for
// the list operation.
type listParams struct {
	fx.In

	ctx     context.Context
	sources *sources.Sources
	logger  common.Logger
}

// listCommand creates and returns the list command that displays all sources.
// It:
// - Sets up the command with appropriate usage and description
// - Configures the command to use runList as its execution function
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured content sources",
		Long: `Display a list of all content sources configured in sources.yml.

Example:
  gocrawl sources list`,
		Run: runList,
	}
}

// runList executes the list command and displays all sources.
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
			if err := executeList(params); err != nil {
				l.Error("Error executing list", "error", err)
				exitCode = 1
			}
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

// executeList retrieves and displays the list of sources.
// It:
// - Gets all sources from the sources instance
// - Handles empty results
// - Displays the sources in a formatted table
func executeList(p *listParams) error {
	if len(p.sources.Sources) == 0 {
		p.logger.Info("No sources found")
		return nil
	}

	return printSources(p.sources.Sources, p.logger)
}

// printSources formats and displays the sources in a table.
// It:
// - Creates a new table with appropriate headers
// - Handles errors gracefully
// - Renders the table with all source information
func printSources(sources []sources.Config, logger common.Logger) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "URL", "Article Index", "Content Index", "Rate Limit", "Max Depth"})

	for _, source := range sources {
		t.AppendRow([]interface{}{
			source.Name,
			source.URL,
			source.ArticleIndex,
			source.Index,
			source.RateLimit,
			source.MaxDepth,
		})
	}

	if t.Length() == 0 {
		logger.Info("No sources found")
		return nil
	}

	t.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(listCommand())
}
