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
	crawlercfg "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	internalsources "github.com/jonesrussell/gocrawl/internal/sources"
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

// Lister handles listing sources
type Lister struct {
	sourceManager internalsources.Interface
	logger        logger.Interface
	renderer      *TableRenderer
}

// NewLister creates a new Lister instance
func NewLister(
	sourceManager internalsources.Interface,
	logger logger.Interface,
	renderer *TableRenderer,
) *Lister {
	return &Lister{
		sourceManager: sourceManager,
		logger:        logger,
		renderer:      renderer,
	}
}

// Start begins the list operation
func (l *Lister) Start(ctx context.Context) error {
	l.logger.Info("Listing sources")

	sources, err := l.sourceManager.ListSources(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sources) == 0 {
		l.logger.Info("No sources configured")
		return nil
	}

	// Render the table
	return l.renderer.RenderTable(sources)
}

// NewListCommand creates a new list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		Long:  `List all content sources configured in the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context")
			}

			// Get config from context
			configValue := cmd.Context().Value(cmdcommon.ConfigKey)
			cfg, ok := configValue.(config.Interface)
			if !ok {
				return errors.New("config not found in context")
			}

			// Ensure crawler config exists to avoid nil dereference in sources loader
			if cfg.GetCrawlerConfig() == nil {
				if concrete, ok := cfg.(*config.Config); ok {
					concrete.Crawler = crawlercfg.New()
				}
			}

			// If the configured source file does not exist, fall back to sources.example.yml
			if concrete, ok := cfg.(*config.Config); ok {
				path := concrete.GetCrawlerConfig().SourceFile
				if path == "" {
					path = "sources.yml"
				}
				if _, statErr := os.Stat(path); statErr != nil {
					alt := "./sources.example.yml"
					if _, altErr := os.Stat(alt); altErr == nil {
						concrete.Crawler.SourceFile = alt
					}
				}
			}

			// Create Fx app with the module
			fxApp := fx.New(
				// Include required modules
				Module,
				internalsources.Module,

				// Provide existing config
				fx.Provide(func() config.Interface { return cfg }),

				// Provide existing logger
				fx.Provide(func() logger.Interface { return log }),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke list command
				fx.Invoke(func(l *Lister) error {
					return l.Start(cmd.Context())
				}),
			)

			// Start application
			if err := fxApp.Start(cmd.Context()); err != nil {
				return err
			}

			// Stop application
			if err := fxApp.Stop(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
