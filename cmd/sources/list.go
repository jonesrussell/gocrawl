// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	signalhandler "github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Constants for table formatting
const (

	// Column width constants
	sourceColumnWidth   = 80
	indexesColumnWidth  = 80
	configColumnWidth   = 25
	labelFormatterWidth = 8
)

// Params holds the dependencies required for the list operation.
type Params struct {
	fx.In
	SourceManager sources.Interface
	Logger        common.Logger
}

// listParams holds the parameters required for listing sources.
type listParams struct {
	ctx           context.Context
	sources       sources.Interface
	logger        common.Logger
	outputFormat  string
	showMetadata  bool
	showSelectors bool
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
	// Create a cancellable context
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Set up signal handling with a no-op logger initially
	handler := signalhandler.NewSignalHandler(logger.NewNoOp())
	cleanup := handler.Setup(ctx)
	defer cleanup()

	// Create channels for error handling and completion
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	// Initialize the Fx application with required modules
	app := fx.New(
		fx.NopLogger,
		Module,
		fx.Invoke(func(p struct {
			fx.In
			Sources sources.Interface
			Logger  common.Logger
			LC      fx.Lifecycle
		}) {
			// Update the signal handler with the real logger
			handler.SetLogger(p.Logger)
			p.LC.Append(fx.Hook{
				OnStart: func(context.Context) error {
					params := &Params{
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
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("error starting application: %w", err)
	}

	// Set up cleanup for graceful shutdown
	handler.SetCleanup(func() {
		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultOperationTimeout)
		defer stopCancel()

		// Stop the application and handle any shutdown errors
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
		}
	})

	// Wait for either:
	// - A signal interrupt (SIGINT/SIGTERM)
	// - Context cancellation
	// - List completion
	// - List error
	var listErr error
	select {
	case listErr = <-errChan:
		// Error already logged in ExecuteList
	case <-doneChan:
		// Success message already printed in ExecuteList
	}

	// Only wait for shutdown signal if there was an error
	if listErr != nil {
		handler.Wait()
	}

	return listErr
}

// ExecuteList retrieves and displays the list of sources.
func ExecuteList(params Params) error {
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
