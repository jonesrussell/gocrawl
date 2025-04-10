package testutils

import (
	"context"

	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/mock"
)

// MockIndexManager implements storagetypes.IndexManager for testing.
type MockIndexManager struct {
	mock.Mock
}

// EnsureIndex implements storagetypes.IndexManager
func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// DeleteIndex implements storagetypes.IndexManager
func (m *MockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// ListIndices implements storagetypes.IndexManager
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

// IndexExists implements storagetypes.IndexManager
func (m *MockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// Ensure MockIndexManager implements storagetypes.IndexManager
var _ storagetypes.IndexManager = (*MockIndexManager)(nil)

// NewMockIndexManager creates a new MockIndexManager instance.
func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{}
}
