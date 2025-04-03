// Package testutils provides test utilities for the API package.
package testutils

import (
	"context"
	"errors"

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
func (m *MockSearchManager) Search(ctx context.Context, index string, query map[string]any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).([]any)
	if !ok {
		return nil, errors.New("invalid search result type")
	}
	return result, nil
}

// Count implements api.SearchManager.
func (m *MockSearchManager) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	result, ok := args.Get(0).(int64)
	if !ok {
		return 0, errors.New("invalid count result type")
	}
	return result, nil
}

// Aggregate implements api.SearchManager.
func (m *MockSearchManager) Aggregate(ctx context.Context, index string, aggs map[string]any) (map[string]any, error) {
	args := m.Called(ctx, index, aggs)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result, ok := args.Get(0).(map[string]any)
	if !ok {
		return nil, errors.New("invalid aggregation result type")
	}
	return result, nil
}

// Close implements api.SearchManager.
func (m *MockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}
