package sources

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
)

// IndexManagerAdapter adapts api.IndexManager to sources.IndexManager
type IndexManagerAdapter struct {
	manager api.IndexManager
}

// NewIndexManagerAdapter creates a new adapter for api.IndexManager
func NewIndexManagerAdapter(manager api.IndexManager) IndexManager {
	return &IndexManagerAdapter{
		manager: manager,
	}
}

// EnsureIndex implements sources.IndexManager interface by providing a default mapping
func (a *IndexManagerAdapter) EnsureIndex(ctx context.Context, indexName string) error {
	// Default mapping for source indices
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "keyword",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	return a.manager.EnsureIndex(ctx, indexName, mapping)
}
