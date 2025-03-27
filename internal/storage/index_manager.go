package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/common"
)

// IndexManager implements the api.IndexManager interface
type IndexManager struct {
	client *elasticsearch.Client
	logger common.Logger
}

// NewIndexManager creates a new IndexManager instance
func NewIndexManager(client *elasticsearch.Client, logger common.Logger) *IndexManager {
	return &IndexManager{
		client: client,
		logger: logger,
	}
}

// CreateIndex creates a new index with the given name and mapping
func (m *IndexManager) CreateIndex(ctx context.Context, name string, mapping any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	res, err := m.client.Indices.Create(
		name,
		m.client.Indices.Create.WithContext(ctx),
		m.client.Indices.Create.WithBody(&buf),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	return nil
}

// EnsureIndex ensures that an index exists with the given name and mapping
func (m *IndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	exists, err := m.IndexExists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		if err := m.CreateIndex(ctx, name, mapping); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// DeleteIndex deletes an index
func (m *IndexManager) DeleteIndex(ctx context.Context, name string) error {
	res, err := m.client.Indices.Delete(
		[]string{name},
		m.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to delete index: %s", res.String())
	}

	return nil
}

// IndexExists checks if an index exists
func (m *IndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	res, err := m.client.Indices.Exists(
		[]string{name},
		m.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	return res.StatusCode == 200, nil
}

// UpdateMapping updates the mapping for an existing index
func (m *IndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
		return fmt.Errorf("failed to encode mapping: %w", err)
	}

	res, err := m.client.Indices.PutMapping(
		[]string{name},
		&buf,
		m.client.Indices.PutMapping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to update mapping: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			m.logger.Error("Error closing response body", "error", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to update mapping: %s", res.String())
	}

	return nil
}
