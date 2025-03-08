// Package elasticsearch provides Elasticsearch-specific implementations of the indexing interfaces.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/indexing"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Manager implements the indexing.Manager interface using Elasticsearch.
type Manager struct {
	client *elasticsearch.Client
	logger logger.Interface
}

// NewManager creates a new Elasticsearch index manager.
func NewManager(client *elasticsearch.Client, logger logger.Interface) indexing.Manager {
	return &Manager{
		client: client,
		logger: logger,
	}
}

// EnsureIndex ensures that an index exists with the given mapping.
func (m *Manager) EnsureIndex(ctx context.Context, name string, mapping interface{}) error {
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
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index %s: status: %s", name, res.Status())
	}

	m.logger.Info("Created index", "index", name)
	return nil
}

// DeleteIndex deletes an index.
func (m *Manager) DeleteIndex(ctx context.Context, name string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", name, err)
	}
	defer res.Body.Close()

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
func (m *Manager) IndexExists(ctx context.Context, name string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client)
	if err != nil {
		return false, fmt.Errorf("failed to check if index %s exists: %w", name, err)
	}
	defer res.Body.Close()

	exists := !res.IsError()
	return exists, nil
}

// UpdateMapping updates the mapping of an existing index.
func (m *Manager) UpdateMapping(ctx context.Context, name string, mapping interface{}) error {
	exists, err := m.IndexExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("index %s not found", name)
	}

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
		return fmt.Errorf("failed to update mapping for index %s: %w", name, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update mapping for index %s: status: %s", name, res.Status())
	}

	m.logger.Info("Updated mapping", "index", name)
	return nil
}
