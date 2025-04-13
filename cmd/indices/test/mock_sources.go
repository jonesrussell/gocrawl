package test

import (
	"context"

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
	val, ok := args.Get(0).([]*sourceutils.SourceConfig)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
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

func (m *MockSources) GetMetrics() sourceutils.SourcesMetrics {
	args := m.Called()
	val, ok := args.Get(0).(sourceutils.SourcesMetrics)
	if !ok {
		return sourceutils.SourcesMetrics{}
	}
	return val
}

func (m *MockSources) FindByName(name string) *sourceutils.SourceConfig {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	val, ok := args.Get(0).(*sourceutils.SourceConfig)
	if !ok {
		return nil
	}
	return val
}

func (m *MockSources) GetSource(ctx context.Context, name string) (*sourceutils.SourceConfig, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	val, ok := args.Get(0).(*sourceutils.SourceConfig)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

func (m *MockSources) GetSources() ([]sourceutils.SourceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	val, ok := args.Get(0).([]sourceutils.SourceConfig)
	if !ok {
		return nil, args.Error(1)
	}
	return val, args.Error(1)
}

func (m *MockSources) CreateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	args := m.Called(ctx, source)
	return args.Error(0)
}
