// Package storage implements the storage layer for the application.
package storage

import (
	"context"
	"errors"
	"fmt"

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
func NewSearchManager(storage types.Interface, log logger.Interface) api.SearchManager {
	return &SearchManager{
		storage: storage,
		logger:  log,
	}
}

// Search performs a search query.
func (m *SearchManager) Search(ctx context.Context, index string, query map[string]any) ([]any, error) {
	result, err := m.storage.Search(ctx, index, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return result, nil
}

// Count returns the number of documents matching a query.
func (m *SearchManager) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	return m.storage.Count(ctx, index, query)
}

// Aggregate performs an aggregation query.
func (m *SearchManager) Aggregate(ctx context.Context, index string, aggs map[string]any) (map[string]any, error) {
	result, err := m.storage.Aggregate(ctx, index, aggs)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}
	if result == nil {
		return nil, errors.New("aggregation result is nil")
	}
	if converted, ok := result.(map[string]any); ok {
		return converted, nil
	}
	return nil, errors.New("invalid aggregation result type")
}

// Close implements api.SearchManager
func (m *SearchManager) Close() error {
	return m.storage.Close()
}
