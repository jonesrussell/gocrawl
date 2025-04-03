// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
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

// CreateModule provides the create command dependencies
var CreateModule = fx.Module("create",
	// Core dependencies
	Module,
)

// CreateParams contains the dependencies for the create command
type CreateParams struct {
	fx.In
	Context context.Context
	Logger  logger.Interface
	Storage storage.Interface
}

// createCommand creates and returns the command for creating an Elasticsearch index.
// It:
// - Sets up the command with appropriate usage and description
// - Handles the creation of a new index with the specified name
// - Provides feedback on the success or failure of the operation
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
			app := fx.New(
				fx.NopLogger,
				CreateModule,
				fx.Provide(
					func() context.Context { return ctx },
				),
				fx.Invoke(func(p CreateParams) error {
					indexName := args[0]

					// Create the index with default mapping
					if err := p.Storage.CreateIndex(p.Context, indexName, DefaultMapping); err != nil {
						return fmt.Errorf("failed to create index %s: %w", indexName, err)
					}

					p.Logger.Info("Successfully created index", "name", indexName)
					return nil
				}),
			)

			// Start the application and return any error
			if err := app.Start(ctx); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			return nil
		},
	}

	return cmd
}
