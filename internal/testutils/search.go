// Package testutils provides shared testing utilities across the application.
package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/stretchr/testify/mock"
)

// MockSearchManager implements SearchManager interface for testing
type MockSearchManager struct {
	mock.Mock
}

func (m *MockSearchManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockSearchManager) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, nil
	}
	return val, nil
}

func (m *MockSearchManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

func (m *MockSearchManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Ensure MockSearchManager implements api.SearchManager
var _ api.SearchManager = (*MockSearchManager)(nil)
