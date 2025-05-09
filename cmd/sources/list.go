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
	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// TableRenderer handles the display of source data in a table format
type TableRenderer struct {
	logger logger.Interface
}

// NewTableRenderer creates a new TableRenderer instance
func NewTableRenderer(logger logger.Interface) *TableRenderer {
	return &TableRenderer{
		logger: logger,
	}
}

// RenderTable formats and displays the sources in a table format
func (r *TableRenderer) RenderTable(sources []*sourceutils.SourceConfig) error {
	// Initialize table writer with stdout as output
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)

	// Add table headers
	t.AppendHeader(table.Row{"Name", "URL", "Max Depth", "Rate Limit", "Content Index", "Article Index"})

	// Process each source
	for _, source := range sources {
		// Add row to table
		t.AppendRow(table.Row{
			source.Name,
			source.URL,
			source.MaxDepth,
			source.RateLimit,
			source.Index,
			source.ArticleIndex,
		})
	}

	// Render the table
	t.Render()
	return nil
}

// ListCommand implements the list command for sources.
type ListCommand struct {
	sourceManager sources.Interface
	logger        logger.Interface
	renderer      *TableRenderer
}

// NewListCommand creates a new list command.
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long:  `List all content sources configured in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get dependencies from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context")
			}

			configValue := cmd.Context().Value(cmdcommon.ConfigKey)
			cfg, ok := configValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context")
			}

			// Create Fx app with the module
			fxApp := fx.New(
				// Include required modules
				sources.Module,

				// Provide existing config
				fx.Provide(func() config.Interface { return cfg }),

				// Provide existing logger
				fx.Provide(func() logger.Interface { return log }),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke list command
				fx.Invoke(func(sourceManager sources.Interface, logger logger.Interface) error {
					listCmd := &ListCommand{
						sourceManager: sourceManager,
						logger:        logger,
						renderer:      NewTableRenderer(logger),
					}
					return listCmd.Run(cmd.Context())
				}),
			)

			// Start the application
			log.Info("Starting application")
			startErr := fxApp.Start(cmd.Context())
			if startErr != nil {
				log.Error("Failed to start application", "error", startErr)
				return fmt.Errorf("failed to start application: %w", startErr)
			}

			// Stop the application
			log.Info("Stopping application")
			stopErr := fxApp.Stop(cmd.Context())
			if stopErr != nil {
				log.Error("Failed to stop application", "error", stopErr)
				return fmt.Errorf("failed to stop application: %w", stopErr)
			}

			return nil
		},
	}

	return cmd
}

// Run executes the list command.
func (c *ListCommand) Run(ctx context.Context) error {
	c.logger.Info("Listing sources")

	sources, err := c.sourceManager.ListSources(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sources) == 0 {
		c.logger.Info("No sources configured")
		return nil
	}

	// Render the table
	return c.renderer.RenderTable(sources)
}
