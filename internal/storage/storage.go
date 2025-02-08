package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
)

// Storage struct to hold the Elasticsearch client
type Storage struct {
	ESClient *elasticsearch.Client
}

// NewStorage initializes a new Storage instance
func NewStorage(esClient *elasticsearch.Client) *Storage {
	return &Storage{ESClient: esClient}
}

// IndexDocument indexes a document in Elasticsearch
func (s *Storage) IndexDocument(index string, docID string, document interface{}) error {
	// Convert the document to JSON
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("Error indexing document ID %s: %s", docID, res.String())
	}

	log.Printf("Indexed document ID %s in index %s", docID, index)
	return nil
}

// Other storage-related methods can be added here
