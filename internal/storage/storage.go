package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/config" // Import the config package
)

// Storage struct to hold the Elasticsearch client
type Storage struct {
	ESClient *elasticsearch.Client
}

// NewStorage initializes a new Storage instance
func NewStorage(cfg *config.Config) (*Storage, error) {
	// Create the Elasticsearch client with authentication
	esURL := cfg.ElasticURL
	esPassword := cfg.ElasticPassword
	esAPIKey := cfg.ElasticAPIKey

	cfgElasticsearch := elasticsearch.Config{
		Addresses: []string{esURL},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Check if API key is provided
	if esAPIKey != "" {
		cfgElasticsearch.APIKey = esAPIKey // Use API key for authentication
	} else {
		cfgElasticsearch.Username = "elastic"  // Default username for Elasticsearch
		cfgElasticsearch.Password = esPassword // Use password for authentication
	}

	esClient, err := elasticsearch.NewClient(cfgElasticsearch)
	if err != nil {
		return nil, err
	}

	return &Storage{ESClient: esClient}, nil
}

// IndexDocument indexes a document in Elasticsearch
func (s *Storage) IndexDocument(index string, docID string, document interface{}) error {
	// Convert the document to JSON
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}

	// Log the document being indexed
	log.Printf("Indexing document ID %s in index %s: %s", docID, index, string(data))

	// Create the request to index the document
	req := bytes.NewReader(data)
	res, err := s.ESClient.Index(
		index,
		req,
		s.ESClient.Index.WithDocumentID(docID),
		s.ESClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID %s: %s", docID, res.String())
	}

	log.Printf("Indexed document ID %s in index %s", docID, index)
	return nil
}

// TestConnection checks the connection to the Elasticsearch cluster
func (s *Storage) TestConnection() error {
	ctx := context.Background()
	info, err := s.ESClient.Info(
		s.ESClient.Info.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("error getting Elasticsearch info: %w", err)
	}
	defer info.Body.Close()

	var esInfo map[string]interface{}
	if err := json.NewDecoder(info.Body).Decode(&esInfo); err != nil {
		return fmt.Errorf("error decoding Elasticsearch info response: %w", err)
	}

	log.Printf("Elasticsearch info: %+v", esInfo)
	return nil
}
