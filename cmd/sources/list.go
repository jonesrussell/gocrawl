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

// ListParams holds the parameters required for listing sources.
type ListParams struct {
	fx.In

	SourceManager sources.Interface `name:"sourceManager"`
	Logger        common.Logger
}

// ListCommand creates and returns the list command that displays all sources.
func ListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured content sources",
		Long: `Display a list of all content sources configured in sources.yml.

Example:
  gocrawl sources list`,
		RunE: RunList,
	}
}

// RunList executes the list command and displays all sources.
func RunList(cmd *cobra.Command, _ []string) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules
	app := fx.New(
		fx.NopLogger,
		Module,
		fx.Invoke(func(p struct {
			fx.In
			Sources sources.Interface `name:"sourceManager"`
			Logger  common.Logger
			LC      fx.Lifecycle
		}) {
			p.LC.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &ListParams{
						SourceManager: p.Sources,
						Logger:        p.Logger,
					}
					if err := ExecuteList(*params); err != nil {
						p.Logger.Error("Error executing list", "error", err)
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
		common.PrintInfof("Received signal %v, initiating shutdown...", sig)
	case <-cmd.Context().Done():
		common.PrintInfof("Context cancelled, initiating shutdown...")
	case listErr = <-errChan:
		// Error already logged in ExecuteList
	case <-doneChan:
		// Success message already printed in ExecuteList
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

// ExecuteList retrieves and displays the list of sources.
func ExecuteList(params ListParams) error {
	allSources := params.SourceManager.GetSources()
	if len(allSources) == 0 {
		params.Logger.Info("No sources found")
		return nil
	}

	return PrintSources(allSources, params.Logger)
}

// PrintSources formats and displays the sources in a table.
func PrintSources(sources []sources.Config, logger common.Logger) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Configure table style
	t.SetStyle(table.StyleRounded)
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateRows = true
	t.Style().Options.SeparateHeader = true

	// Create transformers for consistent formatting
	labelTransformer := text.Transformer(func(val any) string {
		str, ok := val.(string)
		if !ok {
			return fmt.Sprintf("%-*s", labelFormatterWidth, "ERROR")
		}
		return fmt.Sprintf("%-*s", labelFormatterWidth, str)
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
		t.AppendRow([]any{
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
