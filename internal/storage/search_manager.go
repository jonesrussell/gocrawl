// Package storage implements the storage layer for the application.
package storage

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// SearchManager implements the api.SearchManager interface
type SearchManager struct {
	storage types.Interface
	logger  logger.Interface
}

// NewSearchManager creates a new search manager instance
func NewSearchManager(storage types.Interface, logger logger.Interface) api.SearchManager {
	return &SearchManager{
		storage: storage,
		logger:  logger,
	}
}

// Search implements api.SearchManager
func (m *SearchManager) Search(ctx context.Context, index string, query map[string]interface{}) ([]interface{}, error) {
	return m.storage.Search(ctx, index, query)
}

// Count implements api.SearchManager
func (m *SearchManager) Count(ctx context.Context, index string, query map[string]interface{}) (int64, error) {
	return m.storage.Count(ctx, index, query)
}

// Aggregate implements api.SearchManager
func (m *SearchManager) Aggregate(ctx context.Context, index string, aggs map[string]interface{}) (map[string]interface{}, error) {
	result, err := m.storage.Aggregate(ctx, index, aggs)
	if err != nil {
		return nil, err
	}
	return result.(map[string]interface{}), nil
}

// Close implements api.SearchManager
func (m *SearchManager) Close() error {
	return m.storage.Close()
}
