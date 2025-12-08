// Package sources provides the sources command implementation.
package sources

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
)

// SourcesCommand implements the sources command.
type SourcesCommand struct {
	sourceManager sources.Interface
	logger        logger.Interface
}

// NewSourcesCommand creates a new sources command.
func NewSourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage content sources",
		Long:  `Manage content sources for crawling`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get dependencies - NEW WAY
			deps, err := common.NewCommandDeps()
			if err != nil {
				return fmt.Errorf("failed to get dependencies: %w", err)
			}

			// Construct dependencies
			sourceManager, err := sources.LoadSources(deps.Config, deps.Logger)
			if err != nil {
				return fmt.Errorf("failed to load sources: %w", err)
			}

			sourcesCmd := &SourcesCommand{
				sourceManager: sourceManager,
				logger:        deps.Logger,
			}

			return sourcesCmd.Run(cmd.Context())
		},
	}

	// Add subcommands
	cmd.AddCommand(
		NewListCommand(),
		NewGenerateCommand(),
		NewValidateCommand(),
	)

	return cmd
}

// Run executes the sources command.
func (c *SourcesCommand) Run(ctx context.Context) error {
	c.logger.Info("Listing sources")

	sourcesList, err := c.sourceManager.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sourcesList) == 0 {
		c.logger.Info("No sources configured")
		return nil
	}

	// Print sources in a formatted table
	log := c.logger
	log.Info("Configured Sources:")
	log.Info("------------------")
	for i := range sourcesList {
		src := &sourcesList[i]
		log.Info("Source details",
			"name", src.Name,
			"url", src.URL,
			"allowed_domains", src.AllowedDomains,
			"start_urls", src.StartURLs,
			"max_depth", src.MaxDepth,
			"rate_limit", src.RateLimit,
			"index", src.Index,
			"article_index", src.ArticleIndex,
		)
		log.Info("------------------")
	}

	return nil
}
