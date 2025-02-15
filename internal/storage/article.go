package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

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
		return fmt.Errorf("error marshaling mappings: %w", err)
	}

	// Create the index
	req := esapi.IndicesCreateRequest{
		Index: "articles",
		Body:  bytes.NewReader(data),
	}

	es.Logger.Debug("Attempting to create index", "index", "articles", "mappings", mappings)

	res, err := req.Do(ctx, es.ESClient)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			es.Logger.Error("Failed to close response body", "error", err)
		}
	}(res.Body)

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("error parsing error response: %w", err)
		}
		es.Logger.Error("Failed to create index", "error", e["error"])
		return fmt.Errorf("elasticsearch error: %v", e["error"])
	}

	es.Logger.Info("Index created successfully", "index", "articles")
	return nil
}

// IndexArticle indexes a single article
func (es *ElasticsearchStorage) IndexArticle(ctx context.Context, article *models.Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("error marshaling article: %w", err)
	}

	es.Logger.Debug("Indexing article", "articleID", article.ID)

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

	es.Logger.Info("Article indexed successfully", "articleID", article.ID)
	return nil
}

// SearchArticles searches for articles based on a query
func (es *ElasticsearchStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	es.Logger.Debug("Searching articles", "query", query, "size", size)

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^2", "body", "tags"},
			},
		},
		"size": size,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("error encoding search query: %w", err)
	}

	res, err := es.ESClient.Search(
		es.ESClient.Search.WithContext(ctx),
		es.ESClient.Search.WithIndex("articles"),
		es.ESClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Check if "hits" exists and is of the expected type
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error parsing hits: expected map[string]interface{}")
	}

	hitItems, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("error parsing hit items: expected []interface{}")
	}

	articles := make([]*models.Article, 0, len(hitItems))

	for _, hit := range hitItems {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error parsing hit: expected map[string]interface{}")
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("error parsing source: expected map[string]interface{}")
		}

		var article models.Article
		sourceData, err := json.Marshal(source)
		if err != nil {
			return nil, fmt.Errorf("error marshaling hit source: %w", err)
		}

		if err := json.Unmarshal(sourceData, &article); err != nil {
			return nil, fmt.Errorf("error unmarshaling article: %w", err)
		}

		articles = append(articles, &article)
	}

	es.Logger.Info("Search completed", "query", query, "results", len(articles))
	return articles, nil
}
