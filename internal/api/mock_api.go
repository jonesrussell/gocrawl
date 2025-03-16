package api

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"
)

// MockIndexManager is a mock implementation of the IndexManager interface
type MockIndexManager struct {
	mock.Mock
}

// EnsureIndex implements IndexManager
func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// DeleteIndex implements IndexManager
func (m *MockIndexManager) DeleteIndex(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// IndexExists implements IndexManager
func (m *MockIndexManager) IndexExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// UpdateMapping implements IndexManager
func (m *MockIndexManager) UpdateMapping(ctx context.Context, name string, mapping any) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// Aggregate implements IndexManager
func (m *MockIndexManager) Aggregate(ctx context.Context, index string, aggs any) (any, error) {
	args := m.Called(ctx, index, aggs)
	return args.Get(0), args.Error(1)
}

// Count implements IndexManager
func (m *MockIndexManager) Count(ctx context.Context, index string, query any) (int64, error) {
	args := m.Called(ctx, index, query)
	err := args.Error(1)
	if err != nil {
		return 0, err
	}
	val, ok := args.Get(0).(int64)
	if !ok {
		return 0, errors.New("invalid type assertion for count result")
	}
	return val, nil
}

// Search implements IndexManager
func (m *MockIndexManager) Search(ctx context.Context, index string, query any) ([]any, error) {
	args := m.Called(ctx, index, query)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	val, ok := args.Get(0).([]any)
	if !ok {
		return nil, errors.New("invalid type assertion for search result")
	}
	return val, nil
}

// NewMockIndexManager creates a new instance of MockIndexManager with default expectations
func NewMockIndexManager() *MockIndexManager {
	m := &MockIndexManager{}
	// Set up default expectations
	m.On("EnsureIndex", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.On("DeleteIndex", mock.Anything, mock.Anything).Return(nil)
	m.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	m.On("UpdateMapping", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.On("Aggregate", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	m.On("Count", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)
	m.On("Search", mock.Anything, mock.Anything, mock.Anything).Return([]any{}, nil)
	return m
}
