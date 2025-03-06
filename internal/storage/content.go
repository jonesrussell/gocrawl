package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/models"
)

// CreateContentIndex creates the content index with appropriate mappings
func (s *ElasticsearchStorage) CreateContentIndex(ctx context.Context) error {
	// Define the mappings for the content index
	mappings := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type": "text",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"url": map[string]interface{}{
					"type": "keyword",
				},
				"type": map[string]interface{}{
					"type": "keyword",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"keywords": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"section": map[string]interface{}{
					"type": "keyword",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"word_count": map[string]interface{}{
					"type": "integer",
				},
				"metadata": map[string]interface{}{
					"type":    "object",
					"dynamic": true,
					"properties": map[string]interface{}{
						"og_title": map[string]interface{}{
							"type": "text",
						},
						"og_description": map[string]interface{}{
							"type": "text",
						},
						"og_image": map[string]interface{}{
							"type": "keyword",
						},
						"og_url": map[string]interface{}{
							"type": "keyword",
						},
						"canonical_url": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
			},
		},
	}

	// Use the mapping service to ensure the mapping is correct
	if err := s.mappingService.EnsureMapping(ctx, s.opts.IndexName, mappings); err != nil {
		return fmt.Errorf("failed to ensure content index mapping: %w", err)
	}

	s.Logger.Info("Content index mapping ensured", "index", s.opts.IndexName)
	return nil
}

// IndexContent indexes a single content document
func (s *ElasticsearchStorage) IndexContent(id string, content *models.Content) error {
	ctx := context.Background()
	body, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("error marshaling content: %w", err)
	}

	res, err := s.ESClient.Index(
		s.opts.IndexName,
		bytes.NewReader(body),
		s.ESClient.Index.WithContext(ctx),
		s.ESClient.Index.WithDocumentID(id),
		s.ESClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("error indexing content: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing content: %s", res.String())
	}

	s.Logger.Info("Content indexed successfully",
		"url", content.URL,
		"id", id,
		"type", content.Type,
		"index", s.opts.IndexName)
	return nil
}

// GetContent retrieves a content document by ID
func (s *ElasticsearchStorage) GetContent(id string) (*models.Content, error) {
	res, err := s.ESClient.Get(
		"content",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting content: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting content: %s", res.String())
	}

	var content models.Content
	if decodeErr := json.NewDecoder(res.Body).Decode(&content); decodeErr != nil {
		return nil, fmt.Errorf("error decoding content: %w", decodeErr)
	}

	return &content, nil
}

// SearchContent searches for content based on a query
func (s *ElasticsearchStorage) SearchContent(query string) ([]*models.Content, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": query,
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error marshaling search query: %w", err)
	}

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithIndex("content"),
		s.ESClient.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("error searching content: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching content: %s", res.String())
	}

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error parsing search response: %w", decodeErr)
	}

	hitsObj, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format: hits object not found")
	}

	hitsArray, ok := hitsObj["hits"].([]interface{})
	if !ok {
		return nil, errors.New("invalid response format: hits array not found")
	}

	contents := make([]*models.Content, 0, len(hitsArray))

	for _, hit := range hitsArray {
		hitMap, isMap := hit.(map[string]interface{})
		if !isMap {
			continue
		}

		source, isSource := hitMap["_source"].(map[string]interface{})
		if !isSource {
			continue
		}

		sourceData, marshalErr := json.Marshal(source)
		if marshalErr != nil {
			return nil, fmt.Errorf("error marshaling source data: %w", marshalErr)
		}

		var content models.Content
		if unmarshalErr := json.Unmarshal(sourceData, &content); unmarshalErr != nil {
			return nil, fmt.Errorf("error unmarshaling content: %w", unmarshalErr)
		}

		contents = append(contents, &content)
	}

	return contents, nil
}

// DeleteContent deletes a content document by ID
func (s *ElasticsearchStorage) DeleteContent(id string) error {
	ctx := context.Background()
	res, err := s.ESClient.Delete(
		s.opts.IndexName,
		id,
		s.ESClient.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting content: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting content: %s", res.String())
	}

	s.Logger.Info("Content deleted successfully", "id", id)
	return nil
}
