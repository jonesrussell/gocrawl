// Package sources implements the command-line interface for managing content sources
// in GoCrawl. This file contains the implementation of the list command that
// displays all configured sources in a formatted table.
package sources

import (
	"context"
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Params holds the dependencies required for the list operation.
type Params struct {
	fx.In
	SourceManager sources.Interface
	Logger        logger.Interface
}

// ListCommand implements the list command for sources.
type ListCommand struct {
	logger        logger.Interface
	sourceManager sources.Interface
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

			// Create source manager
			sourceManager, err := sources.LoadSources(cfg)
			if err != nil {
				if errors.Is(err, loader.ErrNoSources) {
					log.Info("No sources found in configuration. Please add sources to your config file.")
					return nil
				}
				return fmt.Errorf("failed to load sources: %w", err)
			}

			// Create and run the command
			listCmd := &ListCommand{
				logger:        log,
				sourceManager: sourceManager,
			}
			return listCmd.Run(cmd.Context())
		},
	}

	return cmd
}

// Run executes the list command.
func (c *ListCommand) Run(ctx context.Context) error {
	c.logger.Info("Listing sources")

	sources, err := c.sourceManager.GetSources()
	if err != nil {
		return fmt.Errorf("failed to get sources: %w", err)
	}

	if len(sources) == 0 {
		c.logger.Info("No sources configured")
		return nil
	}

	// Print sources in a formatted table
	log := c.logger
	log.Info("Configured Sources:")
	log.Info("------------------")
	for i := range sources {
		source := &sources[i]
		log.Info("Source details",
			"name", source.Name,
			"url", source.URL,
			"allowed_domains", source.AllowedDomains,
			"start_urls", source.StartURLs,
			"max_depth", source.MaxDepth,
			"rate_limit", source.RateLimit,
			"index", source.Index,
			"article_index", source.ArticleIndex,
		)
		log.Info("------------------")
	}

	return nil
}
