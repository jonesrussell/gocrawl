package storage

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Interface defines the methods for the IndexService
type IndexServiceInterface interface {
	EnsureIndex(ctx context.Context, indexName string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)
	CreateIndex(ctx context.Context, indexName string, mapping map[string]interface{}) error
}

// IndexService implements the IndexService interface
type IndexService struct {
	logger  logger.Interface
	storage Interface
}

// Ensure IndexService implements IndexServiceInterface
var _ IndexServiceInterface = (*IndexService)(nil)

// NewIndexService creates a new IndexService instance
func NewIndexService(logger logger.Interface, storage Interface) IndexServiceInterface {
	return &IndexService{
		logger:  logger,
		storage: storage,
	}
}

// CreateIndex creates a new index with the specified mapping
func (s *IndexService) CreateIndex(ctx context.Context, indexName string, mapping map[string]interface{}) error {
	return s.storage.CreateIndex(ctx, indexName, mapping)
}

// EnsureIndex checks if an index exists and creates it if it does not
func (s *IndexService) EnsureIndex(ctx context.Context, indexName string) error {
	exists, checkErr := s.IndexExists(ctx, indexName)
	if checkErr != nil {
		return fmt.Errorf("failed to check index: %w", checkErr)
	}

	if !exists {
		if createErr := s.createArticleIndex(ctx, indexName); createErr != nil {
			return fmt.Errorf("failed to create index: %w", createErr)
		}
	}

	return nil
}

// createArticleIndex creates a new index with the specified mapping
func (s *IndexService) createArticleIndex(ctx context.Context, indexName string) error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"author": map[string]interface{}{
					"type": "keyword",
				},
				"published_date": map[string]interface{}{
					"type": "date",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	return s.CreateIndex(ctx, indexName, mapping)
}

// IndexExists checks if an index exists
func (s *IndexService) IndexExists(ctx context.Context, indexName string) (bool, error) {
	return s.storage.IndexExists(ctx, indexName)
}
