package storage

import (
	"bytes"
	"context"
	"encoding/json"
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

	// Convert mappings to JSON
	data, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Errorf("error marshaling mappings: %w", err)
	}

	// Create the index
	res, err := s.ESClient.Indices.Create(
		s.opts.IndexName,
		s.ESClient.Indices.Create.WithContext(ctx),
		s.ESClient.Indices.Create.WithBody(bytes.NewReader(data)),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	s.Logger.Info("Content index created successfully", "index", s.opts.IndexName)
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
	ctx := context.Background()
	res, err := s.ESClient.Get(
		s.opts.IndexName,
		id,
		s.ESClient.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting content: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting content: %s", res.String())
	}

	var content models.Content
	if err := json.NewDecoder(res.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("error decoding content: %w", err)
	}

	return &content, nil
}

// SearchContent searches for content based on a query
func (s *ElasticsearchStorage) SearchContent(query string) ([]*models.Content, error) {
	ctx := context.Background()
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"crawl_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^2", "body"},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("error encoding search query: %w", err)
	}

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(s.opts.IndexName),
		s.ESClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	contents := make([]*models.Content, 0, len(hits))

	for _, hit := range hits {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		var content models.Content
		sourceData, err := json.Marshal(source)
		if err != nil {
			return nil, fmt.Errorf("error marshaling hit source: %w", err)
		}

		if err := json.Unmarshal(sourceData, &content); err != nil {
			return nil, fmt.Errorf("error unmarshaling content: %w", err)
		}

		contents = append(contents, &content)
	}

	s.Logger.Info("Content search completed", "query", query, "results", len(contents))
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
