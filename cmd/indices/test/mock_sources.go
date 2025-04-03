package test

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/mock"
)

// MockSources implements sources.Interface for testing
type MockSources struct {
	mock.Mock
}

// Source represents a news source configuration
type Source struct {
	Name         string
	Index        string
	ArticleIndex string
}

func (m *MockSources) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*sourceutils.SourceConfig), args.Error(1)
}

func (m *MockSources) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}

func (m *MockSources) DeleteSource(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockSources) ValidateSource(source *sourceutils.SourceConfig) error {
	args := m.Called(source)
	return args.Error(0)
}

func (m *MockSources) GetMetrics() sources.Metrics {
	args := m.Called()
	return args.Get(0).(sources.Metrics)
}

func (m *MockSources) FindByName(name string) *sourceutils.SourceConfig {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sourceutils.SourceConfig)
}

func (m *MockSources) GetSource(ctx context.Context, name string) (*sourceutils.SourceConfig, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sourceutils.SourceConfig), args.Error(1)
}

func (m *MockSources) GetSources() ([]sourceutils.SourceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sourceutils.SourceConfig), args.Error(1)
}

func (m *MockSources) CreateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}
