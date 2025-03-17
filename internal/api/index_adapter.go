// Package api provides interfaces and adapters for external service integrations.
package api

import (
	"context"
)

// IndexManagerAdapter adapts an IndexManager to provide default mappings
type IndexManagerAdapter struct {
	manager IndexManager
}

// NewIndexManagerAdapter creates a new adapter for IndexManager
func NewIndexManagerAdapter(manager IndexManager) IndexManager {
	return &IndexManagerAdapter{
		manager: manager,
	}
}

// EnsureIndex implements IndexManager interface by providing a default mapping
func (a *IndexManagerAdapter) EnsureIndex(ctx context.Context, indexName string, mapping any) error {
	// If no mapping is provided, use default mapping
	if mapping == nil {
		mapping = getDefaultMapping()
	}

	return a.manager.EnsureIndex(ctx, indexName, mapping)
}

// DeleteIndex implements IndexManager interface
func (a *IndexManagerAdapter) DeleteIndex(ctx context.Context, indexName string) error {
	return a.manager.DeleteIndex(ctx, indexName)
}

// IndexExists implements IndexManager interface
func (a *IndexManagerAdapter) IndexExists(ctx context.Context, indexName string) (bool, error) {
	return a.manager.IndexExists(ctx, indexName)
}

// UpdateMapping implements IndexManager interface
func (a *IndexManagerAdapter) UpdateMapping(ctx context.Context, indexName string, mapping any) error {
	return a.manager.UpdateMapping(ctx, indexName, mapping)
}

// getDefaultMapping returns the default mapping for indices
func getDefaultMapping() map[string]any {
	return map[string]any{
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
				"created_at": map[string]any{
					"type": "date",
				},
			},
		},
	}
}
