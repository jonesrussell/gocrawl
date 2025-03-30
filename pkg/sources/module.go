// Package sources provides source management functionality for the application.
package sources

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/fx"
)

// Module provides the sources module for dependency injection.
var Module = fx.Module("sources",
	fx.Provide(
		NewSources,
	),
)

// sources implements the Interface.
type sources struct {
	logger interface {
		Debug(msg string, fields ...any)
		Info(msg string, fields ...any)
		Warn(msg string, fields ...any)
		Error(msg string, fields ...any)
	}
	sources map[string]*Source
	mu      sync.RWMutex
	metrics Metrics
}

// NewSources creates a new Sources instance.
func NewSources(p Params) (Interface, error) {
	if err := ValidateParams(p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	return &sources{
		logger:  p.Logger,
		sources: make(map[string]*Source),
		metrics: Metrics{
			LastUpdated: time.Now(),
		},
	}, nil
}

// GetSource retrieves a source by name.
func (s *sources) GetSource(ctx context.Context, name string) (*Source, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	source, exists := s.sources[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSourceNotFound, name)
	}

	return source, nil
}

// ListSources retrieves all sources.
func (s *sources) ListSources(ctx context.Context) ([]*Source, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]*Source, 0, len(s.sources))
	for _, source := range s.sources {
		sources = append(sources, source)
	}

	return sources, nil
}

// AddSource adds a new source.
func (s *sources) AddSource(ctx context.Context, source *Source) error {
	if err := s.ValidateSource(source); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sources[source.Name]; exists {
		return fmt.Errorf("%w: %s", ErrSourceExists, source.Name)
	}

	s.sources[source.Name] = source
	s.metrics.SourceCount++
	s.metrics.LastUpdated = time.Now()

	s.logger.Info("Added source", "name", source.Name)
	return nil
}

// UpdateSource updates an existing source.
func (s *sources) UpdateSource(ctx context.Context, source *Source) error {
	if err := s.ValidateSource(source); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sources[source.Name]; !exists {
		return fmt.Errorf("%w: %s", ErrSourceNotFound, source.Name)
	}

	s.sources[source.Name] = source
	s.metrics.LastUpdated = time.Now()

	s.logger.Info("Updated source", "name", source.Name)
	return nil
}

// DeleteSource deletes a source by name.
func (s *sources) DeleteSource(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sources[name]; !exists {
		return fmt.Errorf("%w: %s", ErrSourceNotFound, name)
	}

	delete(s.sources, name)
	s.metrics.SourceCount--
	s.metrics.LastUpdated = time.Now()

	s.logger.Info("Deleted source", "name", name)
	return nil
}

// ValidateSource validates a source configuration.
func (s *sources) ValidateSource(source *Source) error {
	if source == nil {
		return fmt.Errorf("%w: source is nil", ErrInvalidSource)
	}

	if source.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidSource)
	}

	if source.URL == "" {
		return fmt.Errorf("%w: URL is required", ErrInvalidSource)
	}

	if source.MaxDepth < 0 {
		return fmt.Errorf("%w: max_depth must be non-negative", ErrInvalidSource)
	}

	return nil
}

// GetMetrics returns the current metrics.
func (s *sources) GetMetrics() Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}
