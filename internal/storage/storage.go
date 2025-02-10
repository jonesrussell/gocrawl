package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Storage defines the methods that any storage implementation must have
type Storage interface {
	IndexDocument(ctx context.Context, index string, docID string, document interface{}) error
	TestConnection(ctx context.Context) error
	BulkIndex(ctx context.Context, index string, documents []interface{}) error
	Search(ctx context.Context, index string, query map[string]interface{}) ([]map[string]interface{}, error)
	CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error
	DeleteIndex(ctx context.Context, index string) error
	UpdateDocument(ctx context.Context, index string, docID string, update map[string]interface{}) error
	DeleteDocument(ctx context.Context, index string, docID string) error
	ScrollSearch(ctx context.Context, index string, query map[string]interface{}, batchSize int) (<-chan map[string]interface{}, error)
}

// ElasticsearchStorage struct to hold the Elasticsearch client
type ElasticsearchStorage struct {
	ESClient *elasticsearch.Client
	Logger   *logger.CustomLogger
}

// Result holds the dependencies for the storage
type Result struct {
	fx.Out

	Storage Storage // Use the interface type here
}

// Ensure ElasticsearchStorage implements the Storage interface
var _ Storage = (*ElasticsearchStorage)(nil)

// NewStorage initializes a new Storage instance
func NewStorage(cfg *config.Config, log *logger.CustomLogger) (Result, error) {
	// Validate essential configuration parameters
	if cfg.ElasticURL == "" {
		return Result{}, errors.New("ELASTIC_URL is required")
	}

	// Create the Elasticsearch client with authentication
	cfgElasticsearch := elasticsearch.Config{
		Addresses: []string{cfg.ElasticURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				//nolint:gosec // This is a temporary solution to bypass the TLS certificate verification
				InsecureSkipVerify: true,
			},
		},
	}

	// Check if API key is provided
	if cfg.ElasticAPIKey != "" {
		cfgElasticsearch.APIKey = cfg.ElasticAPIKey // Use API key for authentication
	} else {
		cfgElasticsearch.Username = "elastic"           // Default username for Elasticsearch
		cfgElasticsearch.Password = cfg.ElasticPassword // Use password for authentication
	}

	esClient, err := elasticsearch.NewClient(cfgElasticsearch)
	if err != nil {
		return Result{}, err
	}

	return Result{Storage: &ElasticsearchStorage{ESClient: esClient, Logger: log}}, nil
}

// IndexDocument indexes a document in Elasticsearch
func (s *ElasticsearchStorage) IndexDocument(
	ctx context.Context,
	index string,
	docID string,
	document interface{},
) error {
	// Convert the document to JSON
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	// Create the request to index the document
	req := bytes.NewReader(data)
	res, err := s.ESClient.Index(
		index,
		req,
		s.ESClient.Index.WithDocumentID(docID),
		s.ESClient.Index.WithRefresh("true"),
		s.ESClient.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID %s: %s", docID, res.String())
	}

	// Log a concise summary instead of the full document
	s.Logger.Info("Indexed document", docID, index, res.Status())

	return nil
}

// TestConnection checks the connection to the Elasticsearch cluster
func (s *ElasticsearchStorage) TestConnection(ctx context.Context) error {
	info, err := s.ESClient.Info(
		s.ESClient.Info.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error getting Elasticsearch info: %w", err)
	}
	defer info.Body.Close()

	var esInfo map[string]interface{}
	if decodeErr := json.NewDecoder(info.Body).Decode(&esInfo); decodeErr != nil {
		return fmt.Errorf("error decoding Elasticsearch info response: %w", decodeErr)
	}

	s.Logger.Info("Elasticsearch info", esInfo)
	return nil
}

// BulkIndex performs bulk indexing of documents
func (s *ElasticsearchStorage) BulkIndex(ctx context.Context, index string, documents []interface{}) error {
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
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	hits, _ := result["hits"].(map[string]interface{})
	hitsArr, _ := hits["hits"].([]interface{})

	var documents []map[string]interface{}
	for _, hit := range hitsArr {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		documents = append(documents, source)
	}

	return documents, nil
}

// CreateIndex creates a new index with optional mapping
func (s *ElasticsearchStorage) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
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

	return nil
}

// DeleteDocument deletes a document
func (s *ElasticsearchStorage) DeleteDocument(ctx context.Context, index string, docID string) error {
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

// ScrollSearch implements scroll API for handling large result sets
func (s *ElasticsearchStorage) ScrollSearch(
	ctx context.Context,
	index string,
	query map[string]interface{},
	batchSize int,
) (<-chan map[string]interface{}, error) {
	resultChan := make(chan map[string]interface{})

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	go func() {
		defer close(resultChan)

		res, err := s.ESClient.Search(
			s.ESClient.Search.WithContext(ctx),
			s.ESClient.Search.WithIndex(index),
			s.ESClient.Search.WithBody(&buf),
			s.ESClient.Search.WithScroll(time.Minute),
			s.ESClient.Search.WithSize(batchSize),
		)
		if err != nil {
			s.Logger.Error("Initial scroll failed", err)
			return
		}

		for {
			var result map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
				s.Logger.Error("Error parsing scroll response", err)
				res.Body.Close()
				return
			}
			res.Body.Close()

			hits, _ := result["hits"].(map[string]interface{})
			hitsArr, _ := hits["hits"].([]interface{})

			if len(hitsArr) == 0 {
				break
			}

			for _, hit := range hitsArr {
				hitMap := hit.(map[string]interface{})
				select {
				case resultChan <- hitMap["_source"].(map[string]interface{}):
				case <-ctx.Done():
					return
				}
			}

			scrollID := result["_scroll_id"].(string)
			res, err = s.ESClient.Scroll(
				s.ESClient.Scroll.WithScrollID(scrollID),
				s.ESClient.Scroll.WithScroll(time.Minute),
			)
			if err != nil {
				s.Logger.Error("Scroll failed", err)
				return
			}
		}
	}()

	return resultChan, nil
}
