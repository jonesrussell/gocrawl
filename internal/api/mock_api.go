package api

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockIndexManager is a mock implementation of the IndexManager interface
type MockIndexManager struct {
	mock.Mock
}

// EnsureIndex implements IndexManager
func (m *MockIndexManager) EnsureIndex(ctx context.Context, name string, mapping interface{}) error {
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
func (m *MockIndexManager) UpdateMapping(ctx context.Context, name string, mapping interface{}) error {
	args := m.Called(ctx, name, mapping)
	return args.Error(0)
}

// NewMockIndexManager creates a new instance of MockIndexManager with default expectations
func NewMockIndexManager() *MockIndexManager {
	m := &MockIndexManager{}
	// Set up default expectations
	m.On("EnsureIndex", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.On("DeleteIndex", mock.Anything, mock.Anything).Return(nil)
	m.On("IndexExists", mock.Anything, mock.Anything).Return(true, nil)
	m.On("UpdateMapping", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return m
}
