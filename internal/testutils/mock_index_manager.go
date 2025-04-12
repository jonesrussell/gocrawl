package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockIndexManager implements interfaces.IndexManager for testing.
type MockIndexManager struct {
	mock.Mock
}

// EnsureIndex implements interfaces.IndexManager
func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// DeleteIndex implements interfaces.IndexManager
func (m *MockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// ListIndices implements interfaces.IndexManager
func (m *MockIndexManager) ListIndices(ctx context.Context) ([]string, error) {
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

// IndexExists implements interfaces.IndexManager
func (m *MockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// GetMapping implements interfaces.IndexManager
func (m *MockIndexManager) GetMapping(ctx context.Context, name string) (map[string]any, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	mapping, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, args.Error(1)
	}
	return mapping, args.Error(1)
}

// UpdateMapping implements interfaces.IndexManager
func (m *MockIndexManager) UpdateMapping(ctx context.Context, name string, mapping map[string]any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// GetIndex implements interfaces.IndexManager
func (m *MockIndexManager) GetIndex(ctx context.Context, name string) (map[string]any, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	index, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, args.Error(1)
	}
	return index, args.Error(1)
}

// EnsureArticleIndex implements interfaces.IndexManager
func (m *MockIndexManager) EnsureArticleIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// EnsureContentIndex implements interfaces.IndexManager
func (m *MockIndexManager) EnsureContentIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// Ensure MockIndexManager implements interfaces.IndexManager
var _ interfaces.IndexManager = (*MockIndexManager)(nil)

// NewMockIndexManager creates a new mock index manager instance.
func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}
