package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/mock"
)

// MockSourceManager is a mock implementation of sources.Interface
type MockSourceManager struct {
	mock.Mock
}

// NewMockSourceManager creates a new instance of MockSourceManager
func NewMockSourceManager() *MockSourceManager {
	return &MockSourceManager{}
}

// FindByName implements sources.Interface
func (m *MockSourceManager) FindByName(name string) (*sources.Config, error) {
	args := m.Called(name)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, ErrInvalidResult
	}
	if val, ok := result.(*sources.Config); ok {
		return val, nil
	}
	return nil, ErrInvalidResult
}

// GetSources implements sources.Interface
func (m *MockSourceManager) GetSources() ([]sources.Config, error) {
	args := m.Called()
	if err := args.Error(1); err != nil {
		return nil, err
	}
	result := args.Get(0)
	if result == nil {
		return nil, nil
	}
	if val, ok := result.([]sources.Config); ok {
		return val, nil
	}
	return nil, sources.ErrInvalidSource
}

// Validate implements sources.Interface
func (m *MockSourceManager) Validate(source *sources.Config) error {
	args := m.Called(source)
	return args.Error(0)
}

// AddSource implements sources.Interface
func (m *MockSourceManager) AddSource(ctx context.Context, source *sources.Config) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

// DeleteSource implements sources.Interface
func (m *MockSourceManager) DeleteSource(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

// ListSources implements sources.Interface
func (m *MockSourceManager) ListSources(ctx context.Context) ([]*sources.Config, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*sources.Config), args.Error(1)
}

// UpdateSource implements sources.Interface
func (m *MockSourceManager) UpdateSource(ctx context.Context, source *sources.Config) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

// ValidateSource implements sources.Interface
func (m *MockSourceManager) ValidateSource(source *sources.Config) error {
	args := m.Called(source)
	return args.Error(0)
}

// GetMetrics implements sources.Interface
func (m *MockSourceManager) GetMetrics() sources.Metrics {
	args := m.Called()
	result := args.Get(0)
	if result == nil {
		return sources.Metrics{}
	}
	if val, ok := result.(sources.Metrics); ok {
		return val
	}
	return sources.Metrics{}
}

// Ensure MockSourceManager implements sources.Interface
var _ sources.Interface = (*MockSourceManager)(nil)
