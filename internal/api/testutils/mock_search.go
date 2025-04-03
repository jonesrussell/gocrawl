// Package testutils provides test utilities for the API package.
package testutils

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockSearchManager implements api.SearchManager for testing.
type MockSearchManager struct {
	mock.Mock
}

// NewMockSearchManager creates a new mock search manager instance.
func NewMockSearchManager() *MockSearchManager {
	return &MockSearchManager{}
}

// Search implements api.SearchManager.
func (m *MockSearchManager) Search(ctx context.Context, index string, query map[string]interface{}) ([]interface{}, error) {
	args := m.Called(ctx, index, query)
	if result, ok := args.Get(0).([]interface{}); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// Count implements api.SearchManager.
func (m *MockSearchManager) Count(ctx context.Context, index string, query map[string]interface{}) (int64, error) {
	args := m.Called(ctx, index, query)
	if result, ok := args.Get(0).(int64); ok {
		return result, args.Error(1)
	}
	return 0, args.Error(1)
}

// Aggregate implements api.SearchManager.
func (m *MockSearchManager) Aggregate(ctx context.Context, index string, aggs map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, index, aggs)
	if result, ok := args.Get(0).(map[string]interface{}); ok {
		return result, args.Error(1)
	}
	return nil, args.Error(1)
}

// Close implements api.SearchManager.
func (m *MockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
