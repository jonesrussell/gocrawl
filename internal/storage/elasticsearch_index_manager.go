package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// ElasticsearchIndexManager implements the interfaces.IndexManager interface using Elasticsearch.
type ElasticsearchIndexManager struct {
	client *elasticsearch.Client
	logger logger.Interface
}

// NewElasticsearchIndexManager creates a new Elasticsearch index manager.
func NewElasticsearchIndexManager(client *elasticsearch.Client, logger logger.Interface) interfaces.IndexManager {
	return &ElasticsearchIndexManager{
		client: client,
		logger: logger,
	}
}

// EnsureIndex ensures that an index exists with the specified mapping.
func (m *ElasticsearchIndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	exists, err := m.IndexExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		m.logger.Info("Index already exists", "index", name)
		return nil
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: name,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", name, err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to create index %s: status: %s", name, res.Status())
	}

	m.logger.Info("Created index", "index", name)
	return nil
}

// DeleteIndex deletes an index.
func (m *ElasticsearchIndexManager) DeleteIndex(ctx context.Context, name string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", name, err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return fmt.Errorf("index %s not found", name)
		}
		return fmt.Errorf("failed to delete index %s: status: %s", name, res.Status())
	}

	m.logger.Info("Deleted index", "index", name)
	return nil
}

// IndexExists checks if an index exists.
func (m *ElasticsearchIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return false, fmt.Errorf("failed to check if index %s exists: %w", name, err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	exists := !res.IsError()
	return exists, nil
}

// UpdateMapping updates the mapping for an index.
func (m *ElasticsearchIndexManager) UpdateMapping(ctx context.Context, name string, mapping map[string]any) error {
	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesPutMappingRequest{
		Index: []string{name},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update mapping: %s", res.String())
	}

	m.logger.Info("Successfully updated mapping", "index", name)
	return nil
}

// GetMapping gets the mapping for an index.
func (m *ElasticsearchIndexManager) GetMapping(ctx context.Context, name string) (map[string]any, error) {
	req := esapi.IndicesGetMappingRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get mapping: %s", res.String())
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mapping: %w", err)
	}

	// Extract the mapping from the response
	if mappings, ok := result[name].(map[string]any); ok {
		if mapping, ok := mappings["mappings"].(map[string]any); ok {
			return mapping, nil
		}
	}

	return nil, errors.New("unexpected mapping format")
}
