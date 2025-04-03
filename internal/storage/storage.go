package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/mitchellh/mapstructure"
)

// Constants for timeout durations
const (
	DefaultBulkIndexTimeout      = 30 * time.Second
	DefaultIndexTimeout          = 10 * time.Second
	DefaultTestConnectionTimeout = 5 * time.Second
	DefaultSearchTimeout         = 10 * time.Second
)

// Storage implements the storage interface
type Storage struct {
	client *es.Client
	logger logger.Interface
	opts   Options
}

// NewStorage creates a new storage instance
func NewStorage(client *es.Client, logger logger.Interface, opts Options) types.Interface {
	return &Storage{
		client: client,
		logger: logger,
		opts:   opts,
	}
}

// Ensure Storage implements types.Interface
var _ types.Interface = (*Storage)(nil)

// Helper function to create a context with timeout
func (s *Storage) createContextWithTimeout(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// IndexDocument indexes a document in Elasticsearch
func (s *Storage) IndexDocument(ctx context.Context, index, id string, document any) error {
	if s.client == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	body, err := json.Marshal(document)
	if err != nil {
		s.logger.Error("Failed to index document", "error", err)
		return fmt.Errorf("error marshaling document: %w", err)
	}

	res, err := s.client.Index(
		index,
		bytes.NewReader(body),
		s.client.Index.WithContext(ctx),
		s.client.Index.WithDocumentID(id),
		s.client.Index.WithRefresh("true"),
	)
	if err != nil {
		s.logger.Error("Failed to index document", "error", err)
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to index document", "error", res.String())
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	s.logger.Info("Document indexed successfully",
		"index", index,
		"docID", id,
		"type", fmt.Sprintf("%T", document))
	return nil
}

// BulkIndex performs bulk indexing of documents
func (s *Storage) BulkIndex(ctx context.Context, index string, documents []any) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultBulkIndexTimeout)
	defer cancel()

	var buf bytes.Buffer
	if err := s.prepareBulkIndexRequest(&buf, index, documents); err != nil {
		return err
	}

	res, err := s.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		s.client.Bulk.WithContext(ctx),
		s.client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("bulk indexing failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	s.logger.Info("Bulk indexed documents", "count", len(documents), "index", index)
	return nil
}

// prepareBulkIndexRequest prepares the bulk index request
func (s *Storage) prepareBulkIndexRequest(
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
func (s *Storage) Search(ctx context.Context, index string, query any) ([]any, error) {
	if s.client == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	// First check if the index exists
	exists, err := s.IndexExists(ctx, index)
	if err != nil {
		return nil, fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		s.logger.Error("Index not found", "index", index)
		return nil, fmt.Errorf("%w: %s", ErrIndexNotFound, index)
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	body, err := marshalJSON(query)
	if err != nil {
		return nil, fmt.Errorf("error marshaling search query: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(body)),
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

	hits, hasHits := result["hits"].(map[string]any)
	if !hasHits {
		return nil, errors.New("invalid search response format")
	}

	hitsList, hasHitsList := hits["hits"].([]any)
	if !hasHitsList {
		return nil, errors.New("invalid search response format")
	}

	var documents []any
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]any)
		if !ok {
			continue
		}

		source, hasSource := hitMap["_source"]
		if !hasSource {
			continue
		}

		documents = append(documents, source)
	}

	return documents, nil
}

// CreateIndex creates a new index with the specified mapping
func (s *Storage) CreateIndex(
	ctx context.Context,
	index string,
	mapping map[string]any,
) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		s.logger.Error("Failed to create index", "index", index, "error", err)
		return fmt.Errorf("error encoding mapping: %w", err)
	}

	res, err := s.client.Indices.Create(
		index,
		s.client.Indices.Create.WithContext(ctx),
		s.client.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		s.logger.Error("Failed to create index", "index", index, "error", err)
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to create index", "index", index, "error", res.String())
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	s.logger.Info("Created index", "index", index)
	return nil
}

// DeleteIndex deletes an index
func (s *Storage) DeleteIndex(ctx context.Context, index string) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	// Convert single index string to slice of strings
	indices := []string{index}

	res, err := s.client.Indices.Delete(
		indices,
		s.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		s.logger.Error("Failed to delete index", "error", err)
		return fmt.Errorf("error deleting index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to delete index", "error", res.String())
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	s.logger.Info("Deleted index", "index", index)
	return nil
}

// UpdateDocument updates a document in Elasticsearch
func (s *Storage) UpdateDocument(
	ctx context.Context,
	index, docID string,
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

	res, err := s.client.Update(
		index,
		docID,
		&buf,
		s.client.Update.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("update failed: %s", res.String())
	}

	s.logger.Info("Updated document", "index", index, "docID", docID)
	return nil
}

// DeleteDocument deletes a document from Elasticsearch
func (s *Storage) DeleteDocument(ctx context.Context, index, docID string) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	res, err := s.client.Delete(
		index,
		docID,
		s.client.Delete.WithContext(ctx),
	)
	if err != nil {
		s.logger.Error("Failed to delete document", "error", err)
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to delete document", "error", res.String())
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	s.logger.Info("Deleted document", "index", index, "docID", docID)
	return nil
}

// IndexExists checks if the specified index exists
func (s *Storage) IndexExists(ctx context.Context, indexName string) (bool, error) {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultTestConnectionTimeout)
	defer cancel()

	res, err := s.client.Indices.Exists([]string{indexName}, s.client.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	return res.StatusCode == http.StatusOK, nil
}

// GetDocument retrieves a document from Elasticsearch
func (s *Storage) GetDocument(ctx context.Context, index, id string, document any) error {
	res, err := s.client.Get(
		index,
		id,
		s.client.Get.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("error getting document: %s", res.String())
	}

	if decodeErr := json.NewDecoder(res.Body).Decode(document); decodeErr != nil {
		return fmt.Errorf("error decoding document: %w", decodeErr)
	}

	return nil
}

// SearchDocuments searches for documents in Elasticsearch
func (s *Storage) SearchDocuments(
	ctx context.Context,
	index, query string,
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

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("error searching documents: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

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
func (s *Storage) Ping(ctx context.Context) error {
	if s.client == nil {
		return errors.New("elasticsearch client is nil")
	}

	res, err := s.client.Ping(s.client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error pinging Elasticsearch: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("error pinging Elasticsearch: %s", res.String())
	}

	return nil
}

// ListIndices lists all indices in the cluster
func (s *Storage) ListIndices(ctx context.Context) ([]string, error) {
	res, err := s.client.Cat.Indices(
		s.client.Cat.Indices.WithContext(ctx),
		s.client.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		s.logger.Error("Failed to list indices", "error", err)
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to list indices", "error", res.String())
		return nil, fmt.Errorf("error listing indices: %s", res.String())
	}

	var indices []struct {
		Index string `json:"index"`
	}
	if decodeErr := json.NewDecoder(res.Body).Decode(&indices); decodeErr != nil {
		s.logger.Error("Failed to list indices", "error", decodeErr)
		return nil, fmt.Errorf("error decoding indices: %w", decodeErr)
	}

	result := make([]string, len(indices))
	for i, idx := range indices {
		result[i] = idx.Index
	}

	s.logger.Info("Retrieved indices list")
	return result, nil
}

// GetMapping retrieves the mapping for an index
func (s *Storage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	res, err := s.client.Indices.GetMapping(
		s.client.Indices.GetMapping.WithContext(ctx),
		s.client.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		s.logger.Error("Failed to get mapping", "error", err)
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to get mapping", "error", res.String())
		return nil, fmt.Errorf("error getting mapping: %s", res.String())
	}

	var mapping map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&mapping); decodeErr != nil {
		s.logger.Error("Failed to get mapping", "error", decodeErr)
		return nil, fmt.Errorf("error decoding mapping: %w", decodeErr)
	}

	s.logger.Info("Retrieved mapping", "index", index)
	return mapping, nil
}

// UpdateMapping updates the mapping for an index
func (s *Storage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("error encoding mapping: %w", err)
	}

	res, err := s.client.Indices.PutMapping(
		[]string{index},
		&buf,
		s.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("error updating mapping: %s", res.String())
	}

	return nil
}

// GetIndexHealth retrieves the health status of an index
func (s *Storage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	res, err := s.client.Cluster.Health(
		s.client.Cluster.Health.WithContext(ctx),
		s.client.Cluster.Health.WithIndex(index),
	)
	if err != nil {
		s.logger.Error("Failed to get index health", "error", err)
		return "", fmt.Errorf("failed to get index health: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to get index health", "error", res.String())
		return "", fmt.Errorf("error getting index health: %s", res.String())
	}

	var health map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&health); decodeErr != nil {
		s.logger.Error("Failed to get index health", "error", decodeErr)
		return "", fmt.Errorf("error decoding index health: %w", decodeErr)
	}

	status, ok := health["status"].(string)
	if !ok {
		s.logger.Error("Failed to get index health", "error", "invalid index health format")
		return "", ErrInvalidIndexHealth
	}

	s.logger.Info("Retrieved index health", "index", index, "health", status)
	return status, nil
}

// GetIndexDocCount retrieves the document count for an index
func (s *Storage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	res, err := s.client.Count(
		s.client.Count.WithContext(ctx),
		s.client.Count.WithIndex(index),
	)
	if err != nil {
		s.logger.Error("Failed to get index document count", "error", err)
		return 0, fmt.Errorf("failed to get index document count: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to get index document count", "error", res.String())
		return 0, fmt.Errorf("error getting index document count: %s", res.String())
	}

	var count map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&count); decodeErr != nil {
		s.logger.Error("Failed to get index document count", "error", decodeErr)
		return 0, fmt.Errorf("error decoding index document count: %w", decodeErr)
	}

	countValue, ok := count["count"].(float64)
	if !ok {
		s.logger.Error("Failed to get index document count", "error", "invalid index document count format")
		return 0, ErrInvalidDocCount
	}

	s.logger.Info("Retrieved index document count", "index", index, "count", int64(countValue))
	return int64(countValue), nil
}

// marshalJSON marshals the given value to JSON and returns an error if it fails
func marshalJSON(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// SearchArticles searches for articles in Elasticsearch
func (s *Storage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	if s.client == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	body, err := json.Marshal(map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"title^2", "content"},
			},
		},
		"size": size,
	})
	if err != nil {
		s.logger.Error("Failed to marshal search query", "error", err)
		return nil, fmt.Errorf("error marshaling search query: %w", err)
	}

	res, searchErr := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(s.opts.IndexName),
		s.client.Search.WithBody(bytes.NewReader(body)),
	)
	if searchErr != nil {
		s.logger.Error("Failed to execute search", "error", searchErr)
		return nil, fmt.Errorf("error executing search: %w", searchErr)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to execute search", "error", res.String())
		return nil, fmt.Errorf("error executing search: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		s.logger.Error("Failed to decode search result", "error", decodeErr)
		return nil, fmt.Errorf("error decoding search result: %w", decodeErr)
	}

	hits, ok := result["hits"].(map[string]any)
	if !ok {
		s.logger.Error("Failed to decode search result", "error", "invalid search result format")
		return nil, errors.New("invalid search result format")
	}

	hitsArray, ok := hits["hits"].([]any)
	if !ok {
		s.logger.Error("Failed to decode search result", "error", "invalid search result format")
		return nil, errors.New("invalid search result format")
	}

	articles := make([]*models.Article, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, hitOk := hit.(map[string]any)
		if !hitOk {
			continue
		}

		source, sourceOk := hitMap["_source"].(map[string]any)
		if !sourceOk {
			continue
		}

		article := &models.Article{}
		if decodeErr := mapstructure.Decode(source, article); decodeErr != nil {
			s.logger.Error("Failed to decode article", "error", decodeErr)
			continue
		}

		articles = append(articles, article)
	}

	s.logger.Info("Executed search successfully", "query", query, "size", size)
	return articles, nil
}

// Aggregate performs an aggregation query
func (s *Storage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	if s.client == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	body, err := json.Marshal(map[string]any{
		"aggs": aggs,
	})
	if err != nil {
		s.logger.Error("Failed to marshal aggregation query", "error", err)
		return nil, fmt.Errorf("error marshaling aggregation query: %w", err)
	}

	res, searchErr := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
		s.client.Search.WithBody(bytes.NewReader(body)),
	)
	if searchErr != nil {
		s.logger.Error("Failed to execute aggregation", "error", searchErr)
		return nil, fmt.Errorf("error executing aggregation: %w", searchErr)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to execute aggregation", "error", res.String())
		return nil, fmt.Errorf("error executing aggregation: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		s.logger.Error("Failed to decode aggregation result", "error", decodeErr)
		return nil, fmt.Errorf("error decoding aggregation result: %w", decodeErr)
	}

	s.logger.Info("Executed aggregation successfully", "index", index)
	return result["aggregations"], nil
}

// Count counts documents in an index
func (s *Storage) Count(ctx context.Context, index string, query any) (int64, error) {
	if s.client == nil {
		return 0, errors.New("elasticsearch client is not initialized")
	}

	// First check if the index exists
	exists, err := s.IndexExists(ctx, index)
	if err != nil {
		return 0, fmt.Errorf("failed to check index existence: %w", err)
	}
	if !exists {
		s.logger.Error("Index not found", "index", index)
		return 0, fmt.Errorf("%w: %s", ErrIndexNotFound, index)
	}

	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	body, marshalErr := marshalJSON(query)
	if marshalErr != nil {
		return 0, fmt.Errorf("error marshaling count query: %w", marshalErr)
	}

	res, countErr := s.client.Count(
		s.client.Count.WithContext(ctx),
		s.client.Count.WithIndex(index),
		s.client.Count.WithBody(bytes.NewReader(body)),
	)
	if countErr != nil {
		return 0, fmt.Errorf("failed to execute count: %w", countErr)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			s.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		s.logger.Error("Failed to execute count", "error", res.String())
		return 0, fmt.Errorf("error executing count: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		s.logger.Error("Failed to decode count result", "error", decodeErr)
		return 0, fmt.Errorf("error decoding count result: %w", decodeErr)
	}

	count, ok := result["count"].(float64)
	if !ok {
		s.logger.Error("Failed to decode count result", "error", "invalid count result format")
		return 0, errors.New("invalid count result format")
	}

	s.logger.Info("Executed count successfully", "index", index)
	return int64(count), nil
}

// TestConnection tests the connection to the storage backend
func (s *Storage) TestConnection(ctx context.Context) error {
	if s.client == nil {
		return errors.New("elasticsearch client is nil")
	}

	res, err := s.client.Ping(s.client.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error pinging storage: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error pinging storage: %s", res.String())
	}

	return nil
}

// Close closes any resources held by the search manager.
func (s *Storage) Close() error {
	// No resources to close in this implementation
	return nil
}
