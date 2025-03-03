// Package storage provides mock implementations for testing
package storage

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockIndexService is a mock implementation of the IndexServiceInterface
type MockIndexService struct {
	mock.Mock
}

// Ensure MockIndexService implements IndexServiceInterface
var _ IndexServiceInterface = (*MockIndexService)(nil)

// NewMockIndexService creates a new mock index service
func NewMockIndexService() *MockIndexService {
	return &MockIndexService{}
}

// EnsureIndex is a mock method for ensuring an index exists
func (m *MockIndexService) EnsureIndex(ctx context.Context, indexName string) error {
	args := m.Called(ctx, indexName)
	return args.Error(0)
}

// IndexExists is a mock method to check if an index exists
func (m *MockIndexService) IndexExists(ctx context.Context, indexName string) (bool, error) {
	args := m.Called(ctx, indexName)
	return args.Bool(0), args.Error(1)
}

// CreateIndex is a mock method to create an index
func (m *MockIndexService) CreateIndex(ctx context.Context, indexName string, mapping map[string]interface{}) error {
	args := m.Called(ctx, indexName, mapping)
	return args.Error(0)
}
