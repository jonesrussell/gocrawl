package storage

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// CreateArticlesIndex creates the articles index with appropriate mappings
func (es *ElasticsearchStorage) CreateArticlesIndex(ctx context.Context) error {
	// Define the mappings for the articles index
	mappings := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type": "text",
				},
				"body": map[string]interface{}{
					"type": "text",
				},
				"author": map[string]interface{}{
					"type": "keyword",
				},
				"published_date": map[string]interface{}{
					"type": "date",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// Convert mappings to JSON
	data, err := json.Marshal(mappings)
	if err != nil {
		return err
	}

	// Create the index
	req := esapi.IndicesCreateRequest{
		Index: "articles",
		Body:  bytes.NewReader(data),
	}

	_, err = req.Do(ctx, es.ESClient)
	return err
}
