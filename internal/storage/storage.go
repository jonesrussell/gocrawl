package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

// Storage struct to hold the Elasticsearch client
type Storage struct {
	ESClient *elasticsearch.Client
}

// NewStorage initializes a new Storage instance
func NewStorage() (*Storage, error) {
	// Load credentials from environment variables
	esURL := os.Getenv("ELASTIC_URL")
	esPassword := os.Getenv("ELASTIC_PASSWORD")
	esAPIKey := os.Getenv("ELASTIC_API_KEY") // Load API key from environment variables

	// Create the Elasticsearch client with authentication
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}

	// Check if API key is provided
	if esAPIKey != "" {
		cfg.APIKey = esAPIKey // Use API key for authentication
	} else {
		cfg.Username = "elastic"  // Default username for Elasticsearch
		cfg.Password = esPassword // Use password for authentication
	}

	esClient, err := elasticsearch.NewClient(cfg)
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

// Other storage-related methods can be added here
