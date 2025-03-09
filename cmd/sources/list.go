// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	TableWidth = 160

	// Column width constants
	sourceColumnWidth   = 80
	indexesColumnWidth  = 80
	configColumnWidth   = 25
	labelFormatterWidth = 8
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
		RunE: runList,
	}
}

// runList executes the list command and displays all sources.
// It:
// - Sets up signal handling for graceful shutdown
// - Creates channels for error handling and completion
// - Initializes the Fx application with required modules
// - Handles application lifecycle and error cases
// - Displays the sources list in a formatted table
func runList(cmd *cobra.Command, _ []string) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules
	app := fx.New(
		fx.Options(
			common.Module,
			fx.Provide(
				fx.Annotate(
					func() context.Context {
						return cmd.Context()
					},
					fx.ResultTags(`name:"commandContext"`),
				),
			),
		),
		fx.Options(
			sources.Module,
		),
		fx.Invoke(func(lc fx.Lifecycle, s *sources.Sources, l common.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &listParams{
						ctx:     cmd.Context(),
						sources: s,
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

	// Configure table style
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateRows = true
	t.Style().Options.SeparateHeader = true

	// Create transformers for consistent formatting
	labelTransformer := text.Transformer(func(val interface{}) string {
		return fmt.Sprintf("%-*s", labelFormatterWidth, val.(string)) //nolint:errcheck // Sprintf does not return an error
	})

	// Set column configurations to prevent truncation and align content
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:     "Source",
			WidthMax: sourceColumnWidth,
		},
		{
			Name:        "Indexes",
			WidthMax:    indexesColumnWidth,
			Align:       text.AlignLeft,
			Transformer: labelTransformer,
		},
		{
			Name:        "Crawl Config",
			WidthMax:    configColumnWidth,
			Align:       text.AlignLeft,
			Transformer: labelTransformer,
		},
	})

	t.AppendHeader(table.Row{"Source", "Indexes", "Crawl Config"})

	for _, source := range sources {
		sourceInfo := fmt.Sprintf("%s\n%s", source.Name, source.URL)
		indexes := fmt.Sprintf("Articles: %s\nContent:  %s", source.ArticleIndex, source.Index)
		crawlConfig := fmt.Sprintf("Rate:    %s\nDepth:    %d", source.RateLimit, source.MaxDepth)
		t.AppendRow([]interface{}{
			sourceInfo,
			indexes,
			crawlConfig,
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
