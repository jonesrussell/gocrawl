package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/fx"
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
	IndexDocument(ctx context.Context, index string, id string, document interface{}) error
	GetDocument(ctx context.Context, index string, id string, document interface{}) error
	SearchDocuments(ctx context.Context, index string, query string) ([]map[string]interface{}, error)
	DeleteDocument(ctx context.Context, index string, id string) error

	// Bulk operations
	BulkIndex(ctx context.Context, index string, documents []interface{}) error

	// Index management
	CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error
	DeleteIndex(ctx context.Context, index string) error
	ListIndices(ctx context.Context) ([]string, error)
	GetMapping(ctx context.Context, index string) (map[string]interface{}, error)
	UpdateMapping(ctx context.Context, index string, mapping map[string]interface{}) error

	// Common operations
	Close() error
	Ping(ctx context.Context) error

	// Article operations
	SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error)
	Search(ctx context.Context, query string, size int) ([]Article, error)
	IndexArticle(ctx context.Context, article *models.Article) error
	BulkIndexArticles(ctx context.Context, articles []*models.Article) error

	// Index operations
	IndexExists(ctx context.Context, indexName string) (bool, error)
	TestConnection(ctx context.Context) error

	// Content operations
	IndexContent(id string, content *models.Content) error
	GetContent(id string) (*models.Content, error)
	SearchContent(query string) ([]*models.Content, error)
	DeleteContent(id string) error

	// Index health and stats
	GetIndexHealth(ctx context.Context, index string) (string, error)
	GetIndexDocCount(ctx context.Context, index string) (int64, error)
}

// Article represents a document in Elasticsearch
type Article struct {
	Title   string `json:"title" mapstructure:"title"`
	Content string `json:"content" mapstructure:"content"`
	URL     string `json:"url" mapstructure:"url"`
}

// ElasticsearchStorage struct to hold the Elasticsearch client
type ElasticsearchStorage struct {
	ESClient       *elasticsearch.Client
	Logger         logger.Interface
	opts           Options
	mappingService MappingServiceInterface
	IndexName      string
}

// Result holds the dependencies for the storage
type Result struct {
	fx.Out

	Storage        Interface
	IndexService   IndexServiceInterface
	MappingService MappingServiceInterface
}

// Ensure ElasticsearchStorage implements the Storage interface
var _ Interface = (*ElasticsearchStorage)(nil)

// Error definitions
var (
	ErrInvalidIndexHealth = errors.New("invalid index health format")
	ErrInvalidDocCount    = errors.New("invalid index document count format")
)

// Helper function to create a context with timeout
func (s *ElasticsearchStorage) createContextWithTimeout(
	ctx context.Context,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// IndexDocument indexes a document in Elasticsearch
func (s *ElasticsearchStorage) IndexDocument(ctx context.Context, index string, id string, document interface{}) error {
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

// TestConnection checks if the Elasticsearch client is working
func (s *ElasticsearchStorage) TestConnection(ctx context.Context) error {
	if s.ESClient == nil {
		return errors.New("elasticsearch client is nil")
	}

	// Perform a simple request to test the connection
	res, err := s.ESClient.Info(s.ESClient.Info.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}

// BulkIndex performs bulk indexing of documents
func (s *ElasticsearchStorage) BulkIndex(ctx context.Context, index string, documents []interface{}) error {
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
	documents []interface{},
) error {
	for _, doc := range documents {
		action := map[string]interface{}{
			"index": map[string]interface{}{
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
func (s *ElasticsearchStorage) Search(ctx context.Context, query string, size int) ([]Article, error) {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultSearchTimeout)
	defer cancel()

	s.Logger.Debug("Searching articles", "query", query, "size", size)

	// Build the search query
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"weighted_field_search": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "content"},
			},
		},
		"size": size,
	}

	// Convert query to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		s.Logger.Error("Failed to search documents", "error", err)
		return nil, fmt.Errorf("error encoding search query: %w", err)
	}

	// Perform the search request
	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(s.IndexName),
		s.ESClient.Search.WithBody(&buf),
		s.ESClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		s.Logger.Error("Failed to search documents", "error", err)
		return nil, fmt.Errorf("error searching documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to search documents", "error", res.String())
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	var searchResult map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&searchResult); decodeErr != nil {
		s.Logger.Error("Failed to search documents", "error", decodeErr)
		return nil, fmt.Errorf("error parsing search response: %w", decodeErr)
	}

	articles, total, err := s.parseSearchResponse(searchResult)
	if err != nil {
		return nil, err
	}

	s.Logger.Info("Search completed", "query", query, "results", total)
	return articles, nil
}

// parseSearchResponse parses the search response and returns articles
func (s *ElasticsearchStorage) parseSearchResponse(searchResult map[string]interface{}) ([]Article, int64, error) {
	hits, ok := searchResult["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, errors.New("invalid search response format")
	}

	total := int64(0)
	if totalMap, totalOk := hits["total"].(map[string]interface{}); totalOk {
		if value, valueOk := totalMap["value"].(float64); valueOk {
			total = int64(value)
		}
	}

	var articles []Article
	hitsArray, hitsOk := hits["hits"].([]interface{})
	if !hitsOk {
		return nil, 0, errors.New("invalid hits array format")
	}

	for _, hit := range hitsArray {
		hitMap, hitOk := hit.(map[string]interface{})
		if !hitOk {
			continue
		}

		source, sourceOk := hitMap["_source"].(map[string]interface{})
		if !sourceOk {
			continue
		}

		article := Article{}
		if decodeErr := mapstructure.Decode(source, &article); decodeErr != nil {
			return nil, 0, fmt.Errorf("error decoding article: %w", decodeErr)
		}
		articles = append(articles, article)
	}

	return articles, total, nil
}

// CreateIndex creates a new index with optional mapping
func (s *ElasticsearchStorage) CreateIndex(
	ctx context.Context,
	index string,
	mapping map[string]interface{},
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
	update map[string]interface{},
) error {
	ctx, cancel := s.createContextWithTimeout(ctx, DefaultIndexTimeout)
	defer cancel()

	body := map[string]interface{}{
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

// BulkIndexArticles indexes multiple articles in bulk
func (s *ElasticsearchStorage) BulkIndexArticles(ctx context.Context, articles []*models.Article) error {
	var buf bytes.Buffer

	for _, article := range articles {
		// Add metadata action
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": "articles",
				"_id":    article.ID,
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			s.Logger.Error("Failed to bulk index articles", "error", err)
			return fmt.Errorf("error encoding meta: %w", err)
		}

		// Add document
		if err := json.NewEncoder(&buf).Encode(article); err != nil {
			s.Logger.Error("Failed to bulk index articles", "error", err)
			return fmt.Errorf("error encoding article: %w", err)
		}
	}

	s.Logger.Debug("Bulk indexing articles", "count", len(articles))

	res, err := s.ESClient.Bulk(bytes.NewReader(buf.Bytes()),
		s.ESClient.Bulk.WithContext(ctx), // Ensure context is passed
		s.ESClient.Bulk.WithRefresh("true"))
	if err != nil {
		s.Logger.Error("Failed to bulk index articles", "error", err)
		return fmt.Errorf("bulk indexing failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		s.Logger.Error("Failed to bulk index articles", "error", res.String())
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	s.Logger.Info("Bulk indexed documents", "count", len(articles))
	return nil
}

// Close implements Interface
func (s *ElasticsearchStorage) Close() error {
	return nil // Elasticsearch client doesn't need explicit closing
}

// GetDocument implements Interface
func (s *ElasticsearchStorage) GetDocument(ctx context.Context, index string, id string, document interface{}) error {
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
) ([]map[string]interface{}, error) {
	// Basic query string query
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

	documents := make([]map[string]interface{}, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, isHitMap := hit.(map[string]interface{})
		if !isHitMap {
			continue
		}

		source, hasSource := hitMap["_source"]
		if !hasSource {
			continue
		}

		sourceMap, isMap := source.(map[string]interface{})
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
func (s *ElasticsearchStorage) GetMapping(ctx context.Context, index string) (map[string]interface{}, error) {
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

	var mapping map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&mapping); decodeErr != nil {
		s.Logger.Error("Failed to get mapping", "error", decodeErr)
		return nil, fmt.Errorf("error decoding mapping: %w", decodeErr)
	}

	s.Logger.Info("Retrieved mapping", "index", index)
	return mapping, nil
}

// UpdateMapping updates the mapping for an index
func (s *ElasticsearchStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]interface{}) error {
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

	var health map[string]interface{}
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

	var count map[string]interface{}
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
