package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Storage struct to hold the Elasticsearch client
type Storage struct {
	ESClient *elasticsearch.Client
	Logger   *logger.CustomLogger
}

// NewStorage initializes a new Storage instance
func NewStorage(cfg *config.Config, log *logger.CustomLogger) (*Storage, error) {
	// Validate essential configuration parameters
	if cfg.ElasticURL == "" {
		return nil, errors.New("ELASTIC_URL is required")
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
		return nil, err
	}

	return &Storage{ESClient: esClient, Logger: log}, nil
}

// IndexDocument indexes a document in Elasticsearch
func (s *Storage) IndexDocument(ctx context.Context, index string, docID string, document interface{}) error {
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
	s.Logger.Info("Indexed document",
		s.Logger.Field("docID", docID),
		s.Logger.Field("index", index))

	return nil
}

// TestConnection checks the connection to the Elasticsearch cluster
func (s *Storage) TestConnection(ctx context.Context) error {
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

	s.Logger.Info("Elasticsearch info", s.Logger.Field("info", esInfo))
	return nil
}
