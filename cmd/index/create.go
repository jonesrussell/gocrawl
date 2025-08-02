// Package index implements the command-line interface for managing Elasticsearch
// index in GoCrawl. It provides commands for listing, deleting, and managing
// index in the Elasticsearch cluster.
package index

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// DefaultMapping provides a default mapping for new index
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

// CreateParams holds the parameters for the create command
type CreateParams struct {
	ConfigPath string
	IndexName  string
}

// Creator implements the index create command
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
		c.logger.Info("Index already exists", "index", c.index)
		return nil
	}

	// Create the index
	if createErr := c.storage.CreateIndex(ctx, c.index, DefaultMapping); createErr != nil {
		c.logger.Error("Failed to create index", "index", c.index, "error", createErr)
		return fmt.Errorf("failed to create index %s: %w", c.index, createErr)
	}

	c.logger.Info("Successfully created index", "index", c.index)
	return nil
}
