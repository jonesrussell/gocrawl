// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// DefaultMapping provides a default mapping for new indices
var DefaultMapping = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"title": map[string]any{
				"type": "text",
			},
			"content": map[string]any{
				"type": "text",
			},
			"url": map[string]any{
				"type": "keyword",
			},
			"source": map[string]any{
				"type": "keyword",
			},
			"published_at": map[string]any{
				"type": "date",
			},
			"created_at": map[string]any{
				"type": "date",
			},
		},
	},
}

// Creator implements the indices create command
type Creator struct {
	config  config.Interface
	logger  logger.Interface
	storage types.Interface
	index   string
}

// NewCreator creates a new creator instance
func NewCreator(
	config config.Interface,
	logger logger.Interface,
	storage types.Interface,
	index string,
) *Creator {
	return &Creator{
		config:  config,
		logger:  logger,
		storage: storage,
		index:   index,
	}
}

// Start executes the create operation
func (c *Creator) Start(ctx context.Context) error {
	c.logger.Info("Creating index", "index", c.index)

	if err := c.storage.CreateIndex(ctx, c.index, DefaultMapping); err != nil {
		c.logger.Error("Failed to create index", "index", c.index, "error", err)
		return fmt.Errorf("failed to create index %s: %w", c.index, err)
	}

	c.logger.Info("Successfully created index", "index", c.index)
	return nil
}

// createCommand creates and returns the command for creating an Elasticsearch index.
func createCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [index-name]",
		Short: "Create a new Elasticsearch index",
		Long: `Create a new Elasticsearch index.
This command creates a new index in the Elasticsearch cluster with the specified name.
The index will be created with default settings unless overridden by configuration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a cancellable context
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Initialize the Fx application
			fxApp := fx.New(
				fx.NopLogger,
				Module,
				fx.Provide(
					func() context.Context { return ctx },
					func() string { return args[0] },
				),
				fx.Invoke(func(lc fx.Lifecycle, creator *Creator) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							return creator.Start(ctx)
						},
						OnStop: func(context.Context) error {
							return nil
						},
					})
				}),
			)

			// Start the application
			if err := fxApp.Start(ctx); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			// Stop the application
			if err := fxApp.Stop(ctx); err != nil {
				return fmt.Errorf("error stopping application: %w", err)
			}

			return nil
		},
	}

	return cmd
}
