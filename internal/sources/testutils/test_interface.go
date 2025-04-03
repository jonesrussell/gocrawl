package testutils

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

// TestInterface defines the interface for testing source operations.
type TestInterface interface {
	sources.Interface
}

// TestSources implements TestInterface for testing.
type TestSources struct {
	sources []sourceutils.SourceConfig
}

// NewTestSources creates a new TestSources instance.
func NewTestSources(sources []sourceutils.SourceConfig) TestInterface {
	return &TestSources{
		sources: sources,
	}
}

// ListSources retrieves all sources.
func (s *TestSources) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	result := make([]*sourceutils.SourceConfig, len(s.sources))
	for i := range s.sources {
		result[i] = &s.sources[i]
	}
	return result, nil
}

// AddSource adds a new source.
func (s *TestSources) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	s.sources = append(s.sources, *source)
	return nil
}

// UpdateSource updates an existing source.
func (s *TestSources) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	for i := range s.sources {
		if s.sources[i].Name == source.Name {
			s.sources[i] = *source
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

// DeleteSource deletes a source by name.
func (s *TestSources) DeleteSource(ctx context.Context, name string) error {
	for i := range s.sources {
		if s.sources[i].Name == name {
			s.sources = append(s.sources[:i], s.sources[i+1:]...)
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

// ValidateSource validates a source configuration.
func (s *TestSources) ValidateSource(source *sourceutils.SourceConfig) error {
	if source == nil {
		return sources.ErrInvalidSource
	}
	return nil
}

// GetMetrics returns the current metrics.
func (s *TestSources) GetMetrics() sources.Metrics {
	return sources.Metrics{
		SourceCount: int64(len(s.sources)),
	}
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
