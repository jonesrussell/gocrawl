// Package indices implements the command-line interface for managing Elasticsearch
// indices in GoCrawl. It provides commands for listing, deleting, and managing
// indices in the Elasticsearch cluster.
package indices

import (
	"context"
	"errors"
	"fmt"

	cmdcommon "github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
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
	params CreateParams,
) *Creator {
	return &Creator{
		config:  config,
		logger:  logger,
		storage: storage,
		index:   params.IndexName,
	}
}

// Start executes the create operation
func (c *Creator) Start(ctx context.Context) error {
	c.logger.Info("Creating index", "index", c.index)

	// Test storage connection
	if err := c.storage.TestConnection(ctx); err != nil {
		c.logger.Error("Failed to connect to storage", "error", err)
		return fmt.Errorf("failed to connect to storage: %w", err)
	}

	// Check if index already exists
	exists, err := c.storage.IndexExists(ctx, c.index)
	if err != nil {
		c.logger.Error("Failed to check if index exists", "index", c.index, "error", err)
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists {
		return fmt.Errorf("index %s already exists", c.index)
	}

	// Create the index
	if createErr := c.storage.CreateIndex(ctx, c.index, DefaultMapping); createErr != nil {
		c.logger.Error("Failed to create index", "index", c.index, "error", createErr)
		return fmt.Errorf("failed to create index %s: %w", c.index, createErr)
	}

	c.logger.Info("Successfully created index", "index", c.index)
	return nil
}

// CreateParams holds the parameters for the create command
type CreateParams struct {
	ConfigPath string
	IndexName  string
}

// NewCreateCommand creates a new create command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [index-name]",
		Short: "Create a new Elasticsearch index",
		Long: `Create a new Elasticsearch index.
This command creates a new index in the Elasticsearch cluster with the specified name.
The index will be created with default settings unless overridden by configuration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get logger from context
			loggerValue := cmd.Context().Value(cmdcommon.LoggerKey)
			log, ok := loggerValue.(logger.Interface)
			if !ok {
				return errors.New("logger not found in context or invalid type")
			}

			// Get config path from flags
			configPath, _ := cmd.Flags().GetString("config")

			// Create Fx application
			app := fx.New(
				// Include all required modules
				Module,
				storage.Module,

				// Provide config path string
				fx.Provide(func() string { return configPath }),

				// Provide logger
				fx.Provide(func() logger.Interface { return log }),

				// Provide create params
				fx.Provide(func() CreateParams {
					return CreateParams{
						ConfigPath: configPath,
						IndexName:  args[0],
					}
				}),

				// Use custom Fx logger
				fx.WithLogger(func() fxevent.Logger {
					return logger.NewFxLogger(log)
				}),

				// Invoke create command
				fx.Invoke(func(c *Creator) error {
					if err := c.Start(cmd.Context()); err != nil {
						return err
					}
					cmd.Printf("Successfully created index: %s\n", args[0])
					return nil
				}),
			)

			// Start application
			if err := app.Start(context.Background()); err != nil {
				return err
			}

			// Stop application
			if err := app.Stop(context.Background()); err != nil {
				return err
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("config", "c", "config.yaml", "Path to config file")

	return cmd
}
