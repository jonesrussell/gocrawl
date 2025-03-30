package testutils

import (
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
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	if val, ok := result.([]sources.Config); ok {
		return val, args.Error(1)
	}
	return nil, args.Error(1)
}

// Validate implements sources.Interface
func (m *MockSourceManager) Validate(source *sources.Config) error {
	args := m.Called(source)
	return args.Error(0)
}

// Ensure MockSourceManager implements sources.Interface
var _ sources.Interface = (*MockSourceManager)(nil)
