package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/stretchr/testify/mock"
)

// MockIndexManager implements api.IndexManager for testing.
type MockIndexManager struct {
	mock.Mock
}

// EnsureIndex implements api.IndexManager
func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// DeleteIndex implements api.IndexManager
func (m *MockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// IndexExists implements api.IndexManager
func (m *MockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// UpdateMapping implements api.IndexManager
func (m *MockIndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// Aggregate implements api.IndexManager
func (m *MockIndexManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements api.IndexManager
func (m *MockIndexManager) Count(ctx context.Context, index string, query any) (int64, error) {
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

// Search implements api.IndexManager
func (m *MockIndexManager) Search(ctx context.Context, index string, query any) ([]any, error) {
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

// Ensure MockIndexManager implements api.IndexManager
var _ api.IndexManager = (*MockIndexManager)(nil)

// NewMockIndexManager creates a new MockIndexManager instance.
func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}
