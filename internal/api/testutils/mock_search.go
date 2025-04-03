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
func (m *MockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).([]any), args.Error(1)
}

// Count implements api.SearchManager.
func (m *MockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(int64), args.Error(1)
}

// Aggregate implements api.SearchManager.
func (m *MockSearchManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Close implements api.SearchManager.
func (m *MockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
