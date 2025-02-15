package storage

import (
	"context"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/logger"
)

type IndexService struct {
	storage Interface
	logger  logger.Interface
}

func NewIndexService(storage Interface, logger logger.Interface) *IndexService {
	return &IndexService{
		storage: storage,
		logger:  logger,
	}
}

func (s *IndexService) EnsureIndex(ctx context.Context, indexName string) error {
	exists, checkErr := s.storage.IndexExists(ctx, indexName)
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

	return s.storage.CreateIndex(ctx, indexName, mapping)
}
