package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/mitchellh/mapstructure"
)

// ElasticsearchStorage implements the storage interface using Elasticsearch
type ElasticsearchStorage struct {
	client *es.Client
	config *elasticsearch.Config
	logger logger.Interface
}

// NewElasticsearchStorage creates a new Elasticsearch storage instance
func NewElasticsearchStorage(
	client *es.Client,
	cfg *elasticsearch.Config,
	log logger.Interface,
) *ElasticsearchStorage {
	return &ElasticsearchStorage{
		client: client,
		config: cfg,
		logger: log,
	}
}

// GetIndexManager returns the index manager for this storage
func (s *ElasticsearchStorage) GetIndexManager() types.IndexManager {
	return s
}

// IndexDocument indexes a document
func (s *ElasticsearchStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
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

// GetDocument retrieves a document by ID
func (s *ElasticsearchStorage) GetDocument(ctx context.Context, index, id string, result any) error {
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

	var doc struct {
		Source any `json:"_source"`
	}
	if decodeErr := json.NewDecoder(res.Body).Decode(&doc); decodeErr != nil {
		return fmt.Errorf("error decoding response: %w", decodeErr)
	}

	if doc.Source == nil {
		return errors.New("document not found")
	}

	if unmarshalErr := mapstructure.Decode(doc.Source, result); unmarshalErr != nil {
		return fmt.Errorf("error unmarshaling document: %w", unmarshalErr)
	}

	return nil
}

// DeleteDocument deletes a document
func (s *ElasticsearchStorage) DeleteDocument(ctx context.Context, index, id string) error {
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
func (s *ElasticsearchStorage) SearchDocuments(
	ctx context.Context,
	index string,
	query map[string]any,
	result any,
) error {
	queryBytes := mustJSON(query)
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error searching: %s", res.String())
	}

	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if decodeErr := json.NewDecoder(res.Body).Decode(&searchResult); decodeErr != nil {
		return fmt.Errorf("error decoding response: %w", decodeErr)
	}

	hits := make([]any, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		hits = append(hits, hit.Source)
	}

	if unmarshalErr := mapstructure.Decode(hits, result); unmarshalErr != nil {
		return fmt.Errorf("error unmarshaling hits: %w", unmarshalErr)
	}

	return nil
}

// Search performs a search query
func (s *ElasticsearchStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	var result []any
	queryMap, ok := query.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid query type: expected map[string]any, got %T", query)
	}
	if err := s.SearchDocuments(ctx, index, queryMap, &result); err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
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

	var countResult map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&countResult); decodeErr != nil {
		return 0, fmt.Errorf("error decoding response: %w", decodeErr)
	}

	count, ok := countResult["count"].(float64)
	if !ok {
		return 0, errors.New("invalid count result")
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

	var aggResult map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&aggResult); decodeErr != nil {
		return nil, fmt.Errorf("error decoding response: %w", decodeErr)
	}

	aggregations, ok := aggResult["aggregations"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid aggregations result")
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

	const successStatusCode = 200
	return res.StatusCode == successStatusCode, nil
}

// ListIndices lists all indices
func (s *ElasticsearchStorage) ListIndices(ctx context.Context) ([]string, error) {
	res, err := s.client.Indices.Get(
		[]string{"_all"},
		s.client.Indices.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error listing indices: %s", res.String())
	}

	var indices map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&indices); decodeErr != nil {
		return nil, fmt.Errorf("error decoding response: %w", decodeErr)
	}

	var result []string
	for index := range indices {
		result = append(result, index)
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

	var mappingResult map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&mappingResult); decodeErr != nil {
		return nil, fmt.Errorf("error decoding response: %w", decodeErr)
	}

	indexMapping, ok := mappingResult[index].(map[string]any)
	if !ok {
		return nil, errors.New("invalid index mapping")
	}

	mappings, ok := indexMapping["mappings"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid mappings")
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

	var healthResult map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&healthResult); decodeErr != nil {
		return "", fmt.Errorf("error decoding response: %w", decodeErr)
	}

	status, ok := healthResult["status"].(string)
	if !ok {
		return "", errors.New("invalid status")
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

	var countResult map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&countResult); decodeErr != nil {
		return 0, fmt.Errorf("error decoding response: %w", decodeErr)
	}

	count, ok := countResult["count"].(float64)
	if !ok {
		return 0, errors.New("invalid count result")
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
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal to JSON: %v", err))
	}
	return data
}

// EnsureArticleIndex ensures the article index exists with the correct mapping
func (s *ElasticsearchStorage) EnsureArticleIndex(ctx context.Context, name string) error {
	return s.CreateIndex(ctx, name, getArticleMapping())
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
