// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlercfg "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	internalsources "github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
)

// TableRenderer handles the display of source data in a table format
type TableRenderer struct {
	logger logger.Interface
}

// NewTableRenderer creates a new TableRenderer instance
func NewTableRenderer(log logger.Interface) *TableRenderer {
	return &TableRenderer{
		logger: log,
	}
}

// RenderTable formats and displays the sources in a table format
func (r *TableRenderer) RenderTable(sources []*internalsources.Config) error {
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
	log logger.Interface,
	renderer *TableRenderer,
) *Lister {
	return &Lister{
		sourceManager: sourceManager,
		logger:        log,
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
			// Get dependencies - NEW WAY
			deps, err := common.NewCommandDeps()
			if err != nil {
				return fmt.Errorf("failed to get dependencies: %w", err)
			}

			// Ensure crawler config is properly set up
			if setupErr := setupCrawlerConfig(deps.Config); setupErr != nil {
				return setupErr
			}

			// Construct dependencies
			sourceManager, err := internalsources.LoadSources(deps.Config, deps.Logger)
			if err != nil {
				return fmt.Errorf("failed to load sources: %w", err)
			}

			renderer := NewTableRenderer(deps.Logger)
			lister := NewLister(sourceManager, deps.Logger, renderer)

			// Execute the list command
			return lister.Start(cmd.Context())
		},
	}

	return cmd
}

// setupCrawlerConfig ensures crawler config exists and source file is properly configured.
func setupCrawlerConfig(cfg config.Interface) error {
	concrete, ok := cfg.(*config.Config)
	if !ok {
		return nil
	}

	// Ensure crawler config exists to avoid nil dereference in sources loader
	if cfg.GetCrawlerConfig() == nil {
		concrete.Crawler = crawlercfg.New()
	}

	return nil
}
