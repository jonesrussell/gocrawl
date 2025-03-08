// Package elasticsearch provides Elasticsearch-specific implementations of the indexing interfaces.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/indexing/client"
	"github.com/jonesrussell/gocrawl/internal/indexing/errors"
)

// Manager implements the api.IndexManager interface using Elasticsearch.
type Manager struct {
	client *client.Client
	logger common.Logger
}

// NewManager creates a new Elasticsearch-based index manager.
func NewManager(client *client.Client, logger common.Logger) (api.IndexManager, error) {
	return &Manager{
		client: client,
		logger: logger,
	}, nil
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

	res, err := req.Do(ctx, m.client.Client())
	if err != nil {
		return errors.NewIndexError(name, "create", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewIndexError(name, "create", fmt.Errorf("status: %s", res.Status()))
	}

	m.logger.Info("Created index", "index", name)
	return nil
}

// DeleteIndex deletes an index.
func (m *Manager) DeleteIndex(ctx context.Context, name string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client.Client())
	if err != nil {
		return errors.NewIndexError(name, "delete", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return errors.ErrIndexNotFound
		}
		return errors.NewIndexError(name, "delete", fmt.Errorf("status: %s", res.Status()))
	}

	m.logger.Info("Deleted index", "index", name)
	return nil
}

// IndexExists checks if an index exists.
func (m *Manager) IndexExists(ctx context.Context, name string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{name},
	}

	res, err := req.Do(ctx, m.client.Client())
	if err != nil {
		return false, errors.NewIndexError(name, "exists", err)
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
		return errors.ErrIndexNotFound
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesPutMappingRequest{
		Index: []string{name},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, m.client.Client())
	if err != nil {
		return errors.NewIndexError(name, "update_mapping", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewIndexError(name, "update_mapping", fmt.Errorf("status: %s", res.Status()))
	}

	m.logger.Info("Updated mapping", "index", name)
	return nil
}
