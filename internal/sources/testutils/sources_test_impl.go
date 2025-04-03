// Package testutils provides testing utilities for the sources package.
package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

// TestInterface defines the interface for source management operations in tests
type TestInterface interface {
	ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error)
	AddSource(ctx context.Context, source *sourceutils.SourceConfig) error
	UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error
	DeleteSource(ctx context.Context, name string) error
	ValidateSource(source *sourceutils.SourceConfig) error
	GetMetrics() sources.Metrics
	FindByName(name string) (*sourceutils.SourceConfig, error)
	GetSources() ([]sourceutils.SourceConfig, error)
}

// NewTestSources creates a new Sources instance for testing.
// This function is intended for testing purposes only.
func NewTestSources(configs []sourceutils.SourceConfig) TestInterface {
	// Create a new testSources instance
	s := &testSourcesImpl{
		configs: configs,
	}
	return s
}

// testSourcesImpl implements a simplified version of the sources.Interface
type testSourcesImpl struct {
	configs []sourceutils.SourceConfig
}

// SetSources sets the sources
func (s *testSourcesImpl) SetSources(configs []sourceutils.SourceConfig) {
	s.configs = configs
}

// ListSources retrieves all sources
func (s *testSourcesImpl) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	result := make([]*sourceutils.SourceConfig, 0, len(s.configs))
	for i := range s.configs {
		result = append(result, &s.configs[i])
	}
	return result, nil
}

// AddSource adds a new source
func (s *testSourcesImpl) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	s.configs = append(s.configs, *source)
	return nil
}

// UpdateSource updates an existing source
func (s *testSourcesImpl) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	for i := range s.configs {
		if s.configs[i].Name == source.Name {
			s.configs[i] = *source
			return nil
		}
	}
	return nil
}

// DeleteSource deletes a source by name
func (s *testSourcesImpl) DeleteSource(ctx context.Context, name string) error {
	for i := range s.configs {
		if s.configs[i].Name == name {
			s.configs = append(s.configs[:i], s.configs[i+1:]...)
			return nil
		}
	}
	return nil
}

// ValidateSource validates a source configuration
func (s *testSourcesImpl) ValidateSource(source *sourceutils.SourceConfig) error {
	if source == nil {
		return nil
	}

	if source.Name == "" {
		return nil
	}

	if source.URL == "" {
		return nil
	}

	if source.RateLimit <= 0 {
		return nil
	}

	if source.MaxDepth <= 0 {
		return nil
	}

	return nil
}

// GetMetrics returns the current metrics
func (s *testSourcesImpl) GetMetrics() sources.Metrics {
	return sources.Metrics{
		SourceCount: int64(len(s.configs)),
		LastUpdated: time.Now(),
	}
}

// GetSources retrieves all source configurations
func (s *testSourcesImpl) GetSources() ([]sourceutils.SourceConfig, error) {
	return s.configs, nil
}

// FindByName finds a source by name
func (s *testSourcesImpl) FindByName(name string) (*sourceutils.SourceConfig, error) {
	for i := range s.configs {
		if s.configs[i].Name == name {
			return &s.configs[i], nil
		}
	}
	return nil, fmt.Errorf("source not found: %s", name)
}
