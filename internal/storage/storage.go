package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/models"
)

// Constants for timeout durations
const (
	DefaultBulkIndexTimeout      = 30 * time.Second
	DefaultIndexTimeout          = 10 * time.Second
	DefaultTestConnectionTimeout = 5 * time.Second
	DefaultSearchTimeout         = 10 * time.Second
)

// Interface defines the storage operations
type Interface interface {
	// Document operations
	IndexDocument(ctx context.Context, index string, id string, document any) error
	GetDocument(ctx context.Context, index string, id string, document any) error
	DeleteDocument(ctx context.Context, index string, id string) error

	// Bulk operations
	BulkIndex(ctx context.Context, index string, documents []any) error

	// Index management
	CreateIndex(ctx context.Context, index string, mapping map[string]any) error
	DeleteIndex(ctx context.Context, index string) error
	ListIndices(ctx context.Context) ([]string, error)
	GetMapping(ctx context.Context, index string) (map[string]any, error)
	UpdateMapping(ctx context.Context, index string, mapping map[string]any) error
	IndexExists(ctx context.Context, index string) (bool, error)

	// Search operations
	Search(ctx context.Context, index string, query any) ([]any, error)

	// Index health and stats
	GetIndexHealth(ctx context.Context, index string) (string, error)
	GetIndexDocCount(ctx context.Context, index string) (int64, error)

	// Common operations
	Close() error
	Ping(ctx context.Context) error
	TestConnection(ctx context.Context) error

	// New operations
	Aggregate(ctx context.Context, index string, aggs any) (any, error)

	// Count operation
	Count(ctx context.Context, index string, query any) (int64, error)
}

// Helper function to create a context with timeout
func (s *ElasticsearchStorage) createContextWithTimeout(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// IndexDocument indexes a document in Elasticsearch
func (s *ElasticsearchStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	if s.ESClient == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	body, err := json.Marshal(document)
	if err != nil {
		s.Logger.Error("Failed to index document", "error", err)
		return fmt.Errorf("error marshaling document: %w", err)
	}

	res, err := s.ESClient.Index(
		index,
		bytes.NewReader(body),
		s.ESClient.Index.WithContext(ctx),
		s.ESClient.Index.WithDocumentID(id),
		s.ESClient.Index.WithRefresh("true"),
	)
	if err != nil {
		s.Logger.Error("Failed to index document", "error", err)
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to index document", "error", res.String())
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	s.Logger.Info("Document indexed successfully", "index", index, "docID", id)
	return nil
}

// BulkIndex performs bulk indexing of documents
func (s *ElasticsearchStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultBulkIndexTimeout)
	defer cancel()

	var buf bytes.Buffer
	if err := s.prepareBulkIndexRequest(&buf, index, documents); err != nil {
		return err
	}

	res, err := s.ESClient.Bulk(
		bytes.NewReader(buf.Bytes()),
		s.ESClient.Bulk.WithContext(ctx),
		s.ESClient.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("bulk indexing failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	s.Logger.Info("Bulk indexed documents", "count", len(documents), "index", index)
	return nil
}

// prepareBulkIndexRequest prepares the bulk index request
func (s *ElasticsearchStorage) prepareBulkIndexRequest(
	buf *bytes.Buffer,
	index string,
	documents []any,
) error {
	for _, doc := range documents {
		action := map[string]any{
			"index": map[string]any{
				"_index": index,
			},
		}
		if err := json.NewEncoder(buf).Encode(action); err != nil {
			return fmt.Errorf("error encoding action: %w", err)
		}

		if err := json.NewEncoder(buf).Encode(doc); err != nil {
			return fmt.Errorf("error encoding document: %w", err)
		}
	}
	return nil
}

// Search performs a search query
func (s *ElasticsearchStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	if s.ESClient == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	// First check if the index exists
	exists, err := s.IndexExists(ctx, index)
	if err != nil {
		return nil, fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		s.Logger.Error("Index not found", "index", index)
		return nil, fmt.Errorf("%w: %s", ErrIndexNotFound, index)
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(index),
		s.ESClient.Search.WithBody(bytes.NewReader(mustMarshal(query))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search request failed: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error decoding search response: %w", decodeErr)
	}

	hits, exists := result["hits"].(map[string]any)
	if !exists {
		return nil, errors.New("invalid search response format")
	}

	hitsList, exists := hits["hits"].([]any)
	if !exists {
		return nil, errors.New("invalid hits format in search response")
	}

	results := make([]any, 0, len(hitsList))
	for _, hit := range hitsList {
		hitData, isMap := hit.(map[string]any)
		if !isMap {
			continue
		}
		source, isMap := hitData["_source"].(map[string]any)
		if !isMap {
			continue
		}
		results = append(results, source)
	}

	return results, nil
}

// CreateIndex creates a new index with optional mapping
func (s *ElasticsearchStorage) CreateIndex(
	ctx context.Context,
	index string,
	mapping map[string]any,
) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		s.Logger.Error("Failed to create index", "index", index, "error", err)
		return fmt.Errorf("error encoding mapping: %w", err)
	}

	res, err := s.ESClient.Indices.Create(
		index,
		s.ESClient.Indices.Create.WithContext(ctx),
		s.ESClient.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		s.Logger.Error("Failed to create index", "index", index, "error", err)
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to create index", "index", index, "error", res.String())
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	s.Logger.Info("Created index", "index", index)
	return nil
}

// DeleteIndex deletes an index
func (s *ElasticsearchStorage) DeleteIndex(ctx context.Context, index string) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	// Convert single index string to slice of strings
	indices := []string{index}

	res, err := s.ESClient.Indices.Delete(
		indices,
		s.ESClient.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		s.Logger.Error("Failed to delete index", "error", err)
		return fmt.Errorf("error deleting index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to delete index", "error", res.String())
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	s.Logger.Info("Deleted index", "index", index)
	return nil
}

// UpdateDocument updates an existing document
func (s *ElasticsearchStorage) UpdateDocument(
	ctx context.Context,
	index string,
	docID string,
	update map[string]any,
) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	body := map[string]any{
		"doc":           update,
		"doc_as_upsert": true,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("error encoding update: %w", err)
	}

	res, err := s.ESClient.Update(
		index,
		docID,
		&buf,
		s.ESClient.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("update failed: %s", res.String())
	}

	s.Logger.Info("Updated document", "index", index, "docID", docID)
	return nil
}

// DeleteDocument deletes a document
func (s *ElasticsearchStorage) DeleteDocument(ctx context.Context, index string, docID string) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	res, err := s.ESClient.Delete(
		index,
		docID,
		s.ESClient.Delete.WithContext(ctx),
	)
	if err != nil {
		s.Logger.Error("Failed to delete document", "error", err)
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to delete document", "error", res.String())
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	s.Logger.Info("Deleted document", "index", index, "docID", docID)
	return nil
}

// IndexExists checks if the specified index exists
func (s *ElasticsearchStorage) IndexExists(ctx context.Context, indexName string) (bool, error) {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultTestConnectionTimeout)
	defer cancel()

	res, err := s.ESClient.Indices.Exists([]string{indexName}, s.ESClient.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == http.StatusOK, nil
}

// Close implements Interface
func (s *ElasticsearchStorage) Close() error {
	return nil // Elasticsearch client doesn't need explicit closing
}

// GetDocument implements Interface
func (s *ElasticsearchStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	res, err := s.ESClient.Get(
		index,
		id,
		s.ESClient.Get.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error getting document: %s", res.String())
	}

	if decodeErr := json.NewDecoder(res.Body).Decode(document); decodeErr != nil {
		return fmt.Errorf("error decoding document: %w", decodeErr)
	}

	return nil
}

// SearchDocuments implements Interface
func (s *ElasticsearchStorage) SearchDocuments(
	ctx context.Context,
	index string,
	query string,
) ([]map[string]any, error) {
	searchQuery := map[string]any{
		"query": map[string]any{
			"query_string": map[string]any{
				"query": query,
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error marshaling search query: %w", err)
	}

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(index),
		s.ESClient.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("error searching documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error parsing search response: %w", decodeErr)
	}

	hitsObj, ok := result["hits"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid response format: hits object not found")
	}

	hitsArray, ok := hitsObj["hits"].([]any)
	if !ok {
		return nil, errors.New("invalid response format: hits array not found")
	}

	documents := make([]map[string]any, 0, len(hitsArray))
	for _, hit := range hitsArray {
		resultMap, isValidMap := hit.(map[string]any)
		if !isValidMap {
			continue
		}

		source, hasSource := resultMap["_source"]
		if !hasSource {
			continue
		}

		sourceMap, isMap := source.(map[string]any)
		if !isMap {
			continue
		}

		documents = append(documents, sourceMap)
	}

	return documents, nil
}

// Ping implements Interface
func (s *ElasticsearchStorage) Ping(ctx context.Context) error {
	res, err := s.ESClient.Ping(
		s.ESClient.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error pinging Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error pinging Elasticsearch: %s", res.String())
	}

	return nil
}

// ListIndices lists all indices in the cluster
func (s *ElasticsearchStorage) ListIndices(ctx context.Context) ([]string, error) {
	res, err := s.ESClient.Cat.Indices(
		s.ESClient.Cat.Indices.WithContext(ctx),
		s.ESClient.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		s.Logger.Error("Failed to list indices", "error", err)
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to list indices", "error", res.String())
		return nil, fmt.Errorf("error listing indices: %s", res.String())
	}

	var indices []struct {
		Index string `json:"index"`
	}
	if decodeErr := json.NewDecoder(res.Body).Decode(&indices); decodeErr != nil {
		s.Logger.Error("Failed to list indices", "error", decodeErr)
		return nil, fmt.Errorf("error decoding indices: %w", decodeErr)
	}

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = idx.Index
	}

	s.Logger.Info("Retrieved indices list")
	return result, nil
}

// GetMapping gets the mapping for an index
func (s *ElasticsearchStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	res, err := s.ESClient.Indices.GetMapping(
		s.ESClient.Indices.GetMapping.WithContext(ctx),
		s.ESClient.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		s.Logger.Error("Failed to get mapping", "error", err)
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to get mapping", "error", res.String())
		return nil, fmt.Errorf("error getting mapping: %s", res.String())
	}

	var mapping map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&mapping); decodeErr != nil {
		s.Logger.Error("Failed to get mapping", "error", decodeErr)
		return nil, fmt.Errorf("error decoding mapping: %w", decodeErr)
	}

	s.Logger.Info("Retrieved mapping", "index", index)
	return mapping, nil
}

// UpdateMapping updates the mapping for an index
func (s *ElasticsearchStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("error encoding mapping: %w", err)
	}

	res, err := s.ESClient.Indices.PutMapping(
		[]string{index},
		&buf,
		s.ESClient.Indices.PutMapping.WithContext(ctx),
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

// GetIndexHealth gets the health status of an index
func (s *ElasticsearchStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	res, err := s.ESClient.Cluster.Health(
		s.ESClient.Cluster.Health.WithContext(ctx),
		s.ESClient.Cluster.Health.WithIndex(index),
	)
	if err != nil {
		s.Logger.Error("Failed to get index health", "error", err)
		return "", fmt.Errorf("failed to get index health: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to get index health", "error", res.String())
		return "", fmt.Errorf("error getting index health: %s", res.String())
	}

	var health map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&health); decodeErr != nil {
		s.Logger.Error("Failed to get index health", "error", decodeErr)
		return "", fmt.Errorf("error decoding index health: %w", decodeErr)
	}

	status, ok := health["status"].(string)
	if !ok {
		s.Logger.Error("Failed to get index health", "error", "invalid index health format")
		return "", ErrInvalidIndexHealth
	}

	s.Logger.Info("Retrieved index health", "index", index, "health", status)
	return status, nil
}

// GetIndexDocCount gets the document count of an index
func (s *ElasticsearchStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	res, err := s.ESClient.Count(
		s.ESClient.Count.WithContext(ctx),
		s.ESClient.Count.WithIndex(index),
	)
	if err != nil {
		s.Logger.Error("Failed to get index document count", "error", err)
		return 0, fmt.Errorf("failed to get index document count: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to get index document count", "error", res.String())
		return 0, fmt.Errorf("error getting index document count: %s", res.String())
	}

	var count map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&count); decodeErr != nil {
		s.Logger.Error("Failed to get index document count", "error", decodeErr)
		return 0, fmt.Errorf("error decoding index document count: %w", decodeErr)
	}

	countValue, ok := count["count"].(float64)
	if !ok {
		s.Logger.Error("Failed to get index document count", "error", "invalid index document count format")
		return 0, ErrInvalidDocCount
	}

	s.Logger.Info("Retrieved index document count", "index", index, "count", int64(countValue))
	return int64(countValue), nil
}

// mustMarshal marshals the given value to JSON or panics if it fails
func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}

// SearchArticles implements Interface
func (s *ElasticsearchStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	searchQuery := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				"content": query,
			},
		},
		"size": size,
	}

	res, searchErr := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(s.opts.IndexName),
		s.ESClient.Search.WithBody(bytes.NewReader(mustMarshal(searchQuery))),
	)
	if searchErr != nil {
		return nil, fmt.Errorf("failed to execute search: %w", searchErr)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search request failed: %s", res.String())
	}

	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source *models.Article `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if decodeErr := json.NewDecoder(res.Body).Decode(&searchResult); decodeErr != nil {
		return nil, fmt.Errorf("error decoding search response: %w", decodeErr)
	}

	articles := make([]*models.Article, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		if hit.Source != nil {
			articles = append(articles, hit.Source)
		}
	}

	return articles, nil
}

// Aggregate performs an aggregation query
func (s *ElasticsearchStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	if s.ESClient == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	query := map[string]any{
		"size": 0, // We don't need hits for aggregations
		"aggs": aggs,
	}

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(index),
		s.ESClient.Search.WithBody(bytes.NewReader(mustMarshal(query))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregation: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("aggregation request failed: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error decoding aggregation response: %w", decodeErr)
	}

	aggregations, exists := result["aggregations"].(map[string]any)
	if !exists {
		return nil, errors.New("invalid aggregation response format")
	}

	return aggregations, nil
}

// Count returns the number of documents matching a query
func (s *ElasticsearchStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	if s.ESClient == nil {
		return 0, errors.New("elasticsearch client is not initialized")
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	// Extract just the query part if it's a map
	var queryBody = query

	res, err := s.ESClient.Count(
		s.ESClient.Count.WithContext(ctx),
		s.ESClient.Count.WithIndex(index),
		s.ESClient.Count.WithBody(bytes.NewReader(mustMarshal(queryBody))),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return 0, fmt.Errorf("count request failed: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return 0, fmt.Errorf("error decoding count response: %w", decodeErr)
	}

	count, ok := result["count"].(float64)
	if !ok {
		return 0, errors.New("invalid count response format")
	}

	return int64(count), nil
}
