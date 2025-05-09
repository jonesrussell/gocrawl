package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ElasticsearchStorage implements the storage interface using Elasticsearch
type ElasticsearchStorage struct {
	client *es.Client
	config *elasticsearch.Config
	logger logger.Interface
}

// NewElasticsearchStorage creates a new Elasticsearch storage instance
func NewElasticsearchStorage(client *es.Client, config *elasticsearch.Config, logger logger.Interface) *ElasticsearchStorage {
	return &ElasticsearchStorage{
		client: client,
		config: config,
		logger: logger,
	}
}

// GetIndexManager returns the index manager for this storage
func (s *ElasticsearchStorage) GetIndexManager() types.IndexManager {
	return s
}

// IndexDocument indexes a document
func (s *ElasticsearchStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	res, err := s.client.Index(
		index,
		bytes.NewReader(mustJSON(document)),
		s.client.Index.WithContext(ctx),
		s.client.Index.WithDocumentID(id),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

// GetDocument retrieves a document
func (s *ElasticsearchStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	res, err := s.client.Get(
		index,
		id,
		s.client.Get.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error getting document: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	source, ok := result["_source"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid document source")
	}

	bytes, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("error marshaling source: %w", err)
	}

	if err := json.Unmarshal(bytes, document); err != nil {
		return fmt.Errorf("error unmarshaling document: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document
func (s *ElasticsearchStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	res, err := s.client.Delete(
		index,
		id,
		s.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	return nil
}

// SearchDocuments performs a search query
func (s *ElasticsearchStorage) SearchDocuments(ctx context.Context, index string, query map[string]any, result any) error {
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(mustJSON(query))),
	)
	if err != nil {
		return fmt.Errorf("failed to search documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error searching documents: %s", res.String())
	}

	var searchResult map[string]any
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	hits, ok := searchResult["hits"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid search result")
	}

	sources, ok := hits["hits"].([]any)
	if !ok {
		return fmt.Errorf("invalid hits array")
	}

	var documents []map[string]any
	for _, hit := range sources {
		hitMap, ok := hit.(map[string]any)
		if !ok {
			continue
		}
		source, ok := hitMap["_source"].(map[string]any)
		if !ok {
			continue
		}
		documents = append(documents, source)
	}

	bytes, err := json.Marshal(documents)
	if err != nil {
		return fmt.Errorf("error marshaling documents: %w", err)
	}

	if err := json.Unmarshal(bytes, result); err != nil {
		return fmt.Errorf("error unmarshaling result: %w", err)
	}

	return nil
}

// Search performs a search query
func (s *ElasticsearchStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	var result []any
	if err := s.SearchDocuments(ctx, index, query.(map[string]any), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Count returns the number of documents matching a query
func (s *ElasticsearchStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	res, err := s.client.Count(
		s.client.Count.WithContext(ctx),
		s.client.Count.WithIndex(index),
		s.client.Count.WithBody(bytes.NewReader(mustJSON(query))),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("error counting documents: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid count result")
	}

	return int64(count), nil
}

// Aggregate performs an aggregation query
func (s *ElasticsearchStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(mustJSON(map[string]any{
			"aggs": aggs,
			"size": 0,
		}))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error aggregating: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	aggregations, ok := result["aggregations"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid aggregations result")
	}

	return aggregations, nil
}

// CreateIndex creates an index
func (s *ElasticsearchStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	res, err := s.client.Indices.Create(
		index,
		s.client.Indices.Create.WithContext(ctx),
		s.client.Indices.Create.WithBody(bytes.NewReader(mustJSON(mapping))),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

// DeleteIndex deletes an index
func (s *ElasticsearchStorage) DeleteIndex(ctx context.Context, index string) error {
	res, err := s.client.Indices.Delete(
		[]string{index},
		s.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	return nil
}

// IndexExists checks if an index exists
func (s *ElasticsearchStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := s.client.Indices.Exists(
		[]string{index},
		s.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// ListIndices lists all indices
func (s *ElasticsearchStorage) ListIndices(ctx context.Context) ([]string, error) {
	res, err := s.client.Cat.Indices(
		s.client.Cat.Indices.WithContext(ctx),
		s.client.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error listing indices: %s", res.String())
	}

	var indices []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	var result []string
	for _, index := range indices {
		name, ok := index["index"].(string)
		if !ok {
			continue
		}
		result = append(result, name)
	}

	return result, nil
}

// GetMapping gets the mapping for an index
func (s *ElasticsearchStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	res, err := s.client.Indices.GetMapping(
		s.client.Indices.GetMapping.WithContext(ctx),
		s.client.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting mapping: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	indexMapping, ok := result[index].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid index mapping")
	}

	mappings, ok := indexMapping["mappings"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid mappings")
	}

	return mappings, nil
}

// UpdateMapping updates the mapping for an index
func (s *ElasticsearchStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	res, err := s.client.Indices.PutMapping(
		[]string{index},
		bytes.NewReader(mustJSON(mapping)),
		s.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating mapping: %s", res.String())
	}

	return nil
}

// GetIndexHealth gets the health of an index
func (s *ElasticsearchStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	res, err := s.client.Cluster.Health(
		s.client.Cluster.Health.WithContext(ctx),
		s.client.Cluster.Health.WithIndex(index),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get index health: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return "", fmt.Errorf("error getting index health: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	status, ok := result["status"].(string)
	if !ok {
		return "", fmt.Errorf("invalid status")
	}

	return status, nil
}

// GetIndexDocCount gets the document count for an index
func (s *ElasticsearchStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	res, err := s.client.Count(
		s.client.Count.WithContext(ctx),
		s.client.Count.WithIndex(index),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get document count: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("error getting document count: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid count result")
	}

	return int64(count), nil
}

// TestConnection tests the connection to Elasticsearch
func (s *ElasticsearchStorage) TestConnection(ctx context.Context) error {
	res, err := s.client.Info()
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}

// Close closes the Elasticsearch client
func (s *ElasticsearchStorage) Close() error {
	// The Elasticsearch client doesn't have a Close method
	return nil
}

// mustJSON marshals a value to JSON, panicking on error
func mustJSON(v any) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal to JSON: %v", err))
	}
	return bytes
}

// EnsureArticleIndex ensures the article index exists with the correct mapping
func (s *ElasticsearchStorage) EnsureArticleIndex(ctx context.Context, name string) error {
	articleMapping := map[string]any{
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
				"published_at": map[string]any{
					"type": "date",
				},
				"source": map[string]any{
					"type": "keyword",
				},
				"author": map[string]any{
					"type": "keyword",
				},
				"tags": map[string]any{
					"type": "keyword",
				},
				"categories": map[string]any{
					"type": "keyword",
				},
			},
		},
	}
	return s.CreateIndex(ctx, name, articleMapping)
}

// EnsureIndex ensures that an index exists with the specified mapping
func (s *ElasticsearchStorage) EnsureIndex(ctx context.Context, name string, mapping any) error {
	exists, err := s.IndexExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	mappingMap, ok := mapping.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid mapping type: expected map[string]any, got %T", mapping)
	}

	return s.CreateIndex(ctx, name, mappingMap)
}

// EnsurePageIndex ensures the page index exists with the correct mapping
func (s *ElasticsearchStorage) EnsurePageIndex(ctx context.Context, name string) error {
	pageMapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id": map[string]any{
					"type": "keyword",
				},
				"url": map[string]any{
					"type": "keyword",
				},
				"title": map[string]any{
					"type": "text",
				},
				"content": map[string]any{
					"type": "text",
				},
				"description": map[string]any{
					"type": "text",
				},
				"created_at": map[string]any{
					"type": "date",
				},
				"updated_at": map[string]any{
					"type": "date",
				},
				"status": map[string]any{
					"type": "keyword",
				},
			},
		},
	}
	return s.CreateIndex(ctx, name, pageMapping)
}
