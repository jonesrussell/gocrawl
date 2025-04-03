// Package testutils provides testing utilities for the sources package.
package testutils

import (
	"time"

	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

// TestInterface defines the interface for source management operations in tests
type TestInterface interface {
	ListSources(ctx interface{}) ([]*sourceutils.SourceConfig, error)
	AddSource(ctx interface{}, source *sourceutils.SourceConfig) error
	UpdateSource(ctx interface{}, source *sourceutils.SourceConfig) error
	DeleteSource(ctx interface{}, name string) error
	ValidateSource(source *sourceutils.SourceConfig) error
	GetMetrics() interface{}
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
func (s *testSourcesImpl) ListSources(ctx interface{}) ([]*sourceutils.SourceConfig, error) {
	result := make([]*sourceutils.SourceConfig, 0, len(s.configs))
	for i := range s.configs {
		result = append(result, &s.configs[i])
	}
	return result, nil
}

// AddSource adds a new source
func (s *testSourcesImpl) AddSource(ctx interface{}, source *sourceutils.SourceConfig) error {
	if source == nil {
		return nil
	}

	for _, existing := range s.configs {
		if existing.Name == source.Name {
			return nil
		}
	}

	s.configs = append(s.configs, *source)
	return nil
}

// UpdateSource updates an existing source
func (s *testSourcesImpl) UpdateSource(ctx interface{}, source *sourceutils.SourceConfig) error {
	if source == nil {
		return nil
	}

	for i, existing := range s.configs {
		if existing.Name == source.Name {
			s.configs[i] = *source
			return nil
		}
	}

	return nil
}

// DeleteSource deletes a source by name
func (s *testSourcesImpl) DeleteSource(ctx interface{}, name string) error {
	for i, existing := range s.configs {
		if existing.Name == name {
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
func (s *testSourcesImpl) GetMetrics() interface{} {
	return struct {
		SourceCount int64
		LastUpdated interface{}
	}{
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
	for i, source := range s.configs {
		if source.Name == name {
			return &s.configs[i], nil
		}
	}
	return nil, nil
}
