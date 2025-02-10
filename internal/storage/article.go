package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/models"
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

// IndexArticle indexes a single article
func (es *ElasticsearchStorage) IndexArticle(ctx context.Context, article *models.Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("error marshaling article: %w", err)
	}

	res, err := es.ESClient.Index(
		"articles",
		bytes.NewReader(data),
		es.ESClient.Index.WithContext(ctx),
		es.ESClient.Index.WithDocumentID(article.ID),
	)
	if err != nil {
		return fmt.Errorf("error indexing article: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing article: %s", res.String())
	}

	return nil
}

// BulkIndexArticles indexes multiple articles in bulk
func (es *ElasticsearchStorage) BulkIndexArticles(ctx context.Context, articles []*models.Article) error {
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

	res, err := es.ESClient.Bulk(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("error bulk indexing: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error bulk indexing: %s", res.String())
	}

	return nil
}
