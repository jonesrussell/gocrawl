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
	"go.uber.org/fx"
)

// Constants for timeout durations
const (
	DefaultIndexTimeout          = 10 * time.Second
	DefaultBulkIndexTimeout      = 30 * time.Second
	DefaultTestConnectionTimeout = 5 * time.Second
)

// Interface defines the methods that any storage implementation must have
type Interface interface {
	IndexDocument(ctx context.Context, index string, docID string, document interface{}) error
	TestConnection(ctx context.Context) error
	BulkIndex(ctx context.Context, index string, documents []interface{}) error
	Search(
		ctx context.Context,
		index string,
		query map[string]interface{},
	) ([]map[string]interface{}, error)
	CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error
	DeleteIndex(ctx context.Context, index string) error
	UpdateDocument(
		ctx context.Context,
		index string,
		docID string,
		update map[string]interface{},
	) error
	DeleteDocument(ctx context.Context, index string, docID string) error
	BulkIndexArticles(ctx context.Context, articles []*models.Article) error
	SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error)
	IndexExists(ctx context.Context, indexName string) (bool, error)
}

// ElasticsearchStorage struct to hold the Elasticsearch client
type ElasticsearchStorage struct {
	ESClient *elasticsearch.Client
	Logger   logger.Interface
	opts     Options
}

// Result holds the dependencies for the storage
type Result struct {
	fx.Out

	Storage Interface // Use the interface type here
}

// Ensure ElasticsearchStorage implements the Storage interface
var _ Interface = (*ElasticsearchStorage)(nil)

// Constants for common values
const (
	defaultRefreshValue = "true"
)

// IndexDocument indexes a document in Elasticsearch
func (s *ElasticsearchStorage) IndexDocument(
	ctx context.Context,
	index string,
	docID string,
	document interface{},
) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
	defer cancel()

	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := bytes.NewReader(data)
	res, err := s.ESClient.Index(
		index,
		req,
		s.ESClient.Index.WithDocumentID(docID),
		s.ESClient.Index.WithRefresh(defaultRefreshValue),
		s.ESClient.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID %s: %s", docID, res.String())
	}

	s.Logger.Info("Indexed document",
		"document_id", docID,
		"index", index,
		"status", res.Status(),
	)

	return nil
}

// TestConnection checks if the Elasticsearch client is working
func (s *ElasticsearchStorage) TestConnection(ctx context.Context) error {
	if s.ESClient == nil {
		return errors.New("elasticsearch client is nil")
	}

	// Perform a simple request to test the connection
	res, err := s.ESClient.Info()
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
	ctx, cancel := context.WithTimeout(ctx, DefaultBulkIndexTimeout) // Use constant
	defer cancel()

	var buf bytes.Buffer

	for _, doc := range documents {
		// Add metadata action
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		if err := json.NewEncoder(&buf).Encode(action); err != nil {
			return fmt.Errorf("error encoding action: %w", err)
		}

		// Add document
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return fmt.Errorf("error encoding document: %w", err)
		}
	}

	res, err := s.ESClient.Bulk(bytes.NewReader(buf.Bytes()),
		s.ESClient.Bulk.WithContext(ctx),
		s.ESClient.Bulk.WithRefresh("true"))
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

// Search performs a search query
func (s *ElasticsearchStorage) Search(
	ctx context.Context,
	index string,
	query map[string]interface{},
) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex(index),
		s.ESClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error parsing response: %w", decodeErr)
	}

	hits, err := s.getHitsFromResult(result)
	if err != nil {
		return nil, err
	}

	var documents []map[string]interface{}
	for _, hit := range hits {
		hitMap, isMap := hit.(map[string]interface{})
		if !isMap {
			continue
		}
		source, isSource := hitMap["_source"].(map[string]interface{})
		if !isSource {
			continue
		}
		documents = append(documents, source)
	}

	return documents, nil
}

// CreateIndex creates a new index with optional mapping
func (s *ElasticsearchStorage) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("error encoding mapping: %w", err)
	}

	res, err := s.ESClient.Indices.Create(
		index,
		s.ESClient.Indices.Create.WithContext(ctx),
		s.ESClient.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	s.Logger.Info("Created index", "index", index)
	return nil
}

// DeleteIndex deletes an index
func (s *ElasticsearchStorage) DeleteIndex(ctx context.Context, index string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
	defer cancel()

	// Convert single index string to slice of strings
	indices := []string{index}

	res, err := s.ESClient.Indices.Delete(
		indices,
		s.ESClient.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
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
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
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
	ctx, cancel := context.WithTimeout(ctx, DefaultIndexTimeout) // Use constant
	defer cancel()

	res, err := s.ESClient.Delete(
		index,
		docID,
		s.ESClient.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	s.Logger.Info("Deleted document", "index", index, "docID", docID)
	return nil
}

// IndexExists checks if the specified index exists
func (s *ElasticsearchStorage) IndexExists(ctx context.Context, indexName string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTestConnectionTimeout) // Use constant
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
			return fmt.Errorf("error encoding meta: %w", err)
		}

		// Add document
		if err := json.NewEncoder(&buf).Encode(article); err != nil {
			return fmt.Errorf("error encoding article: %w", err)
		}
	}

	s.Logger.Debug("Bulk indexing articles", "count", len(articles))

	res, err := s.ESClient.Bulk(bytes.NewReader(buf.Bytes()),
		s.ESClient.Bulk.WithContext(ctx), // Ensure context is passed
		s.ESClient.Bulk.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("bulk indexing failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk indexing failed: %s", res.String())
	}

	s.Logger.Info("Bulk indexed documents", "count", len(articles))
	return nil
}
