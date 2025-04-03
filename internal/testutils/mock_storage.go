package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/stretchr/testify/mock"
)

// MockStorage implements storage.Interface for testing.
type MockStorage struct {
	mock.Mock
}

// IndexDocument implements storage.Interface
func (m *MockStorage) IndexDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// GetDocument implements storage.Interface
func (m *MockStorage) GetDocument(ctx context.Context, index, id string, document any) error {
	args := m.Called(ctx, index, id, document)
	return args.Error(0)
}

// DeleteDocument implements storage.Interface
func (m *MockStorage) DeleteDocument(ctx context.Context, index, id string) error {
	args := m.Called(ctx, index, id)
	return args.Error(0)
}

// BulkIndex implements storage.Interface
func (m *MockStorage) BulkIndex(ctx context.Context, index string, documents []any) error {
	args := m.Called(ctx, index, documents)
	return args.Error(0)
}

// CreateIndex implements storage.Interface
func (m *MockStorage) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// DeleteIndex implements storage.Interface
func (m *MockStorage) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

// ListIndices implements storage.Interface
func (m *MockStorage) ListIndices(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, ErrInvalidResult
	}
	if val, ok := result.([]string); ok {
		return val, nil
	}
	return nil, ErrInvalidResult
}

// GetMapping implements storage.Interface
func (m *MockStorage) GetMapping(ctx context.Context, index string) (map[string]any, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, ErrInvalidResult
	}
	if val, ok := result.(map[string]any); ok {
		return val, nil
	}
	return nil, ErrInvalidResult
}

// UpdateMapping implements storage.Interface
func (m *MockStorage) UpdateMapping(ctx context.Context, index string, mapping map[string]any) error {
	args := m.Called(ctx, index, mapping)
	return args.Error(0)
}

// IndexExists implements storage.Interface
func (m *MockStorage) IndexExists(ctx context.Context, index string) (bool, error) {
	args := m.Called(ctx, index)
	return args.Bool(0), args.Error(1)
}

// Search implements storage.Interface
func (m *MockStorage) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, ErrInvalidResult
	}
	if val, ok := result.([]any); ok {
		return val, nil
	}
	return nil, ErrInvalidResult
}

// GetIndexHealth implements storage.Interface
func (m *MockStorage) GetIndexHealth(ctx context.Context, index string) (string, error) {
	args := m.Called(ctx, index)
	return args.String(0), args.Error(1)
}

// GetIndexDocCount implements storage.Interface
func (m *MockStorage) GetIndexDocCount(ctx context.Context, index string) (int64, error) {
	args := m.Called(ctx, index)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	result := args.Get(0)
	if result == nil {
		return 0, ErrInvalidResult
	}
	if val, ok := result.(int64); ok {
		return val, nil
	}
	return 0, ErrInvalidResult
}

// Ping implements storage.Interface
func (m *MockStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestConnection implements storage.Interface
func (m *MockStorage) TestConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Aggregate implements storage.Interface
func (m *MockStorage) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements storage.Interface
func (m *MockStorage) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	result := args.Get(0)
	if result == nil {
		return 0, ErrInvalidResult
	}
	if val, ok := result.(int64); ok {
		return val, nil
	}
	return 0, ErrInvalidResult
}

// Close implements storage.Interface
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockStorage implements storage.Interface
var _ storage.Interface = (*MockStorage)(nil)

// NewMockStorage creates a new MockStorage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{}
}
