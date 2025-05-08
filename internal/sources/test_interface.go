// Package sources provides source management functionality.
package sources

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/mock"
)

// TestSources implements the Interface interface.
type TestSources struct {
	mock.Mock
	sources []sourceutils.SourceConfig
}

// NewTestSources creates a new TestSources instance.
func NewTestSources(sources []sourceutils.SourceConfig) *TestSources {
	return &TestSources{
		sources: sources,
	}
}

// GetSource implements sources.Interface.
func (s *TestSources) GetSource(ctx context.Context, name string) (*sourceutils.SourceConfig, error) {
	args := s.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return &s.sources[0], args.Error(1)
}

// AddSource implements sources.Interface.
func (s *TestSources) AddSource(ctx context.Context, source sourceutils.SourceConfig) error {
	args := s.Called(ctx, source)
	return args.Error(0)
}

// UpdateSource implements sources.Interface.
func (s *TestSources) UpdateSource(ctx context.Context, source sourceutils.SourceConfig) error {
	args := s.Called(ctx, source)
	return args.Error(0)
}

// DeleteSource implements sources.Interface.
func (s *TestSources) DeleteSource(ctx context.Context, name string) error {
	args := s.Called(ctx, name)
	return args.Error(0)
}

// ListSources implements sources.Interface.
func (s *TestSources) ListSources(ctx context.Context) ([]sourceutils.SourceConfig, error) {
	args := s.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return s.sources, args.Error(1)
}

// ValidateSource implements sources.Interface.
func (s *TestSources) ValidateSource(source sourceutils.SourceConfig) error {
	args := s.Called(source)
	return args.Error(0)
}

// GetMetrics implements sources.Interface.
func (s *TestSources) GetMetrics() sourceutils.SourcesMetrics {
	args := s.Called()
	if args.Get(0) == nil {
		return sourceutils.SourcesMetrics{}
	}
	if metrics, ok := args.Get(0).(sourceutils.SourcesMetrics); ok {
		return metrics
	}
	return sourceutils.SourcesMetrics{}
}

// FindByName finds a source by name.
func (s *TestSources) FindByName(name string) *sourceutils.SourceConfig {
	for i := range s.sources {
		if s.sources[i].Name == name {
			return &s.sources[i]
		}
	}
	return nil
}

// GetSources retrieves all source configurations.
func (s *TestSources) GetSources() ([]sourceutils.SourceConfig, error) {
	return s.sources, nil
}

// MockSources is a mock implementation of the sources interface.
type MockSources struct {
	mock.Mock
}

// GetMetrics implements sources.Interface.
func (m *MockSources) GetMetrics() sourceutils.SourcesMetrics {
	args := m.Called()
	if args.Get(0) == nil {
		return sourceutils.SourcesMetrics{}
	}
	if metrics, ok := args.Get(0).(sourceutils.SourcesMetrics); ok {
		return metrics
	}
	return sourceutils.SourcesMetrics{}
}
