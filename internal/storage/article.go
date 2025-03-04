package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// CreateArticlesIndex creates the articles index with appropriate mappings
func (s *ElasticsearchStorage) CreateArticlesIndex(ctx context.Context) error {
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
				"byline_name": map[string]interface{}{
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
				"intro": map[string]interface{}{
					"type": "text",
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"og_title": map[string]interface{}{
					"type": "text",
				},
				"og_description": map[string]interface{}{
					"type": "text",
				},
				"og_image": map[string]interface{}{
					"type": "keyword",
				},
				"og_url": map[string]interface{}{
					"type": "keyword",
				},
				"canonical_url": map[string]interface{}{
					"type": "keyword",
				},
				"word_count": map[string]interface{}{
					"type": "integer",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"section": map[string]interface{}{
					"type": "keyword",
				},
				"keywords": map[string]interface{}{
					"type": "keyword",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
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

	s.Logger.Debug("Attempting to create index", "index", "articles", "mappings", mappings)

	res, err := req.Do(ctx, s.ESClient)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			s.Logger.Error("Failed to close response body", "error", closeErr)
		}
	}(res.Body)

	if res.IsError() {
		var e map[string]interface{}
		if decodeErr := json.NewDecoder(res.Body).Decode(&e); decodeErr != nil {
			return fmt.Errorf("error parsing error response: %w", decodeErr)
		}
		s.Logger.Error("Failed to create index", "error", e["error"])
		return fmt.Errorf("elasticsearch error: %v", e["error"])
	}

	s.Logger.Info("Index created successfully", "index", "articles")
	return nil
}

// IndexArticle indexes a single article
func (s *ElasticsearchStorage) IndexArticle(ctx context.Context, article *models.Article) error {
	data, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("error marshaling article: %w", err)
	}

	s.Logger.Debug("Indexing article", "articleID", article.ID)

	res, err := s.ESClient.Index(
		"articles",
		bytes.NewReader(data),
		s.ESClient.Index.WithContext(ctx),
		s.ESClient.Index.WithDocumentID(article.ID),
	)
	if err != nil {
		return fmt.Errorf("error indexing article: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing article: %s", res.String())
	}

	s.Logger.Info("Article indexed successfully", "articleID", article.ID)
	return nil
}

// SearchArticles searches for articles based on a query
func (s *ElasticsearchStorage) SearchArticles(ctx context.Context, query string, size int) ([]*models.Article, error) {
	s.Logger.Debug("Searching articles", "query", query, "size", size)

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"crawl_match": map[string]interface{}{
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

	res, err := s.ESClient.Search(
		s.ESClient.Search.WithContext(ctx),
		s.ESClient.Search.WithIndex("articles"),
		s.ESClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, fmt.Errorf("error parsing response: %w", decodeErr)
	}

	// Check if "hits" exists and is of the expected type
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, errors.New("error parsing hits: expected map[string]interface{}")
	}

	hitItems, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, errors.New("error parsing hit items: expected []interface{}")
	}

	articles := make([]*models.Article, 0, len(hitItems))

	for _, hit := range hitItems {
		hitMap, hitOk := hit.(map[string]interface{})
		if !hitOk {
			return nil, errors.New("error parsing hit: expected map[string]interface{}")
		}

		source, sourceOk := hitMap["_source"].(map[string]interface{})
		if !sourceOk {
			return nil, errors.New("error parsing source: expected map[string]interface{}")
		}

		var article models.Article
		sourceData, marshalErr := json.Marshal(source)
		if marshalErr != nil {
			return nil, fmt.Errorf("error marshaling hit source: %w", marshalErr)
		}

		if unmarshalErr := json.Unmarshal(sourceData, &article); unmarshalErr != nil {
			return nil, fmt.Errorf("error unmarshaling article: %w", unmarshalErr)
		}

		articles = append(articles, &article)
	}

	s.Logger.Info("Search completed", "query", query, "results", len(articles))
	return articles, nil
}
