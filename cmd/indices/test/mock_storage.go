// Package test provides test utilities for the indices command.
package test

import (
	"context"

	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

// MockStorage implements storage.Interface for testing
type MockStorage struct {
	mock.Mock
	storagetypes.Interface
}

func (m *MockStorage) CreateIndex(ctx context.Context, name string, mapping map[string]any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

func (m *MockStorage) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]string)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockStorage) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	if err := args.Error(1); err != nil {
		return false, err
	}
	val, ok := args.Get(0).(bool)
	if !ok {
		return false, nil
	}
	return val, nil
}

func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) GetIndexHealth(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	if err := args.Error(1); err != nil {
		return "", err
	}
	val, ok := args.Get(0).(string)
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *MockStorage) GetIndexDocCount(ctx context.Context, name string) (int64, error) {
	args := m.Called(ctx, name)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

// Required interface methods
func (m *MockStorage) IndexDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) GetDocument(ctx context.Context, index string, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

func (m *MockStorage) DeleteDocument(ctx context.Context, index string, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]any), args.Error(1)
}

func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}
