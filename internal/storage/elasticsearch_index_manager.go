// Package storage provides Elasticsearch storage implementation.
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

const (
	// HTTPStatusOK is the HTTP status code for successful requests.
	HTTPStatusOK = http.StatusOK
)

// ElasticsearchIndexManager implements the types.IndexManager interface using Elasticsearch.
type ElasticsearchIndexManager struct {
	client *elasticsearch.Client
	logger logger.Interface
}

// NewElasticsearchIndexManager creates a new Elasticsearch index manager.
func NewElasticsearchIndexManager(client *elasticsearch.Client, logger logger.Interface) types.IndexManager {
	return &ElasticsearchIndexManager{
		client: client,
		logger: logger,
	}
}

// EnsureIndex ensures an index exists with the given name and mapping.
func (m *ElasticsearchIndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	exists, err := m.IndexExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	mappingBytes, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	res, err := m.client.Indices.Create(
		name,
		m.client.Indices.Create.WithBody(strings.NewReader(string(mappingBytes))),
		m.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	return nil
}

// DeleteIndex deletes an index with the given name.
func (m *ElasticsearchIndexManager) DeleteIndex(ctx context.Context, name string) error {
	res, err := m.client.Indices.Delete(
		[]string{name},
		m.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	return nil
}

// IndexExists checks if an index exists.
func (m *ElasticsearchIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	res, err := m.client.Indices.Exists(
		[]string{name},
		m.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == HTTPStatusOK, nil
}

// UpdateMapping updates the mapping for an index.
func (m *ElasticsearchIndexManager) UpdateMapping(ctx context.Context, name string, mapping map[string]any) error {
	mappingBytes, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	res, err := m.client.Indices.PutMapping(
		[]string{name},
		strings.NewReader(string(mappingBytes)),
		m.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating mapping: %s", res.String())
	}

	return nil
}

// GetMapping gets the mapping for an index.
func (m *ElasticsearchIndexManager) GetMapping(ctx context.Context, name string) (map[string]any, error) {
	res, err := m.client.Indices.GetMapping(
		m.client.Indices.GetMapping.WithIndex(name),
		m.client.Indices.GetMapping.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error getting mapping: %s", res.String())
	}

	var result map[string]any
	if decodeErr := json.NewDecoder(res.Body).Decode(&result); decodeErr != nil {
		return nil, decodeErr
	}

	return result, nil
}

// EnsureArticleIndex ensures the article index exists.
func (m *ElasticsearchIndexManager) EnsureArticleIndex(ctx context.Context, name string) error {
	articleMapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"title": map[string]any{
					"type": "text",
				},
				"content": map[string]any{
					"type": "text",
				},
				"url": map[string]any{
					"type": "keyword",
				},
				"published_at": map[string]any{
					"type": "date",
				},
				"source": map[string]any{
					"type": "keyword",
				},
				"author": map[string]any{
					"type": "keyword",
				},
				"tags": map[string]any{
					"type": "keyword",
				},
				"categories": map[string]any{
					"type": "keyword",
				},
			},
		},
	}
	return m.EnsureIndex(ctx, name, articleMapping)
}

// EnsurePageIndex ensures the page index exists.
func (m *ElasticsearchIndexManager) EnsurePageIndex(ctx context.Context, name string) error {
	pageMapping := map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id": map[string]any{
					"type": "keyword",
				},
				"url": map[string]any{
					"type": "keyword",
				},
				"title": map[string]any{
					"type": "text",
				},
				"content": map[string]any{
					"type": "text",
				},
				"description": map[string]any{
					"type": "text",
				},
				"keywords": map[string]any{
					"type": "keyword",
				},
				"og_title": map[string]any{
					"type": "text",
				},
				"og_description": map[string]any{
					"type": "text",
				},
				"og_image": map[string]any{
					"type": "keyword",
				},
				"og_url": map[string]any{
					"type": "keyword",
				},
				"canonical_url": map[string]any{
					"type": "keyword",
				},
				"created_at": map[string]any{
					"type": "date",
				},
				"updated_at": map[string]any{
					"type": "date",
				},
			},
		},
	}
	return m.EnsureIndex(ctx, name, pageMapping)
}
