// Package testutils provides test helpers for the sources package.
package testutils

import (
	"context"
	"errors"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/zap"
)

const (
	// DefaultArticleIndex is the default index name for articles
	DefaultArticleIndex = "articles"
	// DefaultContentIndex is the default index name for content
	DefaultContentIndex = "content"
	// ErrInvalidSourceName is the error message for invalid source name
	ErrInvalidSourceName = "source name is required"
	// ErrInvalidSourceURL is the error message for invalid source URL
	ErrInvalidSourceURL = "source URL is required"
	// ErrInvalidRateLimit is the error message for invalid rate limit
	ErrInvalidRateLimit = "rate limit must be positive"
	// ErrInvalidMaxDepth is the error message for invalid max depth
	ErrInvalidMaxDepth = "max depth must be positive"
)

type zapWrapper struct {
	logger *zap.Logger
}

func (w *zapWrapper) Debug(msg string, fields ...any) {
	w.logger.Debug(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Error(msg string, fields ...any) {
	w.logger.Error(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Info(msg string, fields ...any) {
	w.logger.Info(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Warn(msg string, fields ...any) {
	w.logger.Warn(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Fatal(msg string, fields ...any) {
	w.logger.Fatal(msg, zap.Any("fields", fields))
}

func (w *zapWrapper) Printf(format string, args ...any) {
	w.logger.Info(format, zap.Any("args", args))
}

func (w *zapWrapper) Errorf(format string, args ...any) {
	w.logger.Error(format, zap.Any("args", args))
}

func (w *zapWrapper) Sync() error {
	return w.logger.Sync()
}

// NewTestLogger creates a new test logger.
func NewTestLogger() sources.Logger {
	logger, _ := zap.NewDevelopment()
	return &zapWrapper{logger: logger}
}

// NewTestInterface creates a new test sources interface implementation.
func NewTestInterface(configs []sources.Config) sources.Interface {
	// Create a copy of the configs slice to avoid modifying the original
	result := make([]sources.Config, len(configs))
	copy(result, configs)

	// Set default index names for any config that doesn't have them
	for i := range result {
		if result[i].ArticleIndex == "" {
			result[i].ArticleIndex = DefaultArticleIndex
		}
		if result[i].Index == "" {
			result[i].Index = DefaultContentIndex
		}
	}

	return &testSources{
		configs: result,
		logger:  NewTestLogger(),
	}
}

type testSources struct {
	configs []sources.Config
	logger  sources.Logger
}

// ListSources retrieves all sources.
func (s *testSources) ListSources(ctx context.Context) ([]*sources.Config, error) {
	result := make([]*sources.Config, len(s.configs))
	for i := range s.configs {
		result[i] = &s.configs[i]
	}
	return result, nil
}

// AddSource adds a new source.
func (s *testSources) AddSource(ctx context.Context, source *sources.Config) error {
	if source == nil {
		return sources.ErrInvalidSource
	}
	if err := s.ValidateSource(source); err != nil {
		return err
	}

	// Check for duplicate source
	for _, config := range s.configs {
		if config.Name == source.Name {
			return sources.ErrSourceExists
		}
	}

	// Set default index names if empty
	if source.ArticleIndex == "" {
		source.ArticleIndex = DefaultArticleIndex
	}
	if source.Index == "" {
		source.Index = DefaultContentIndex
	}

	s.configs = append(s.configs, *source)
	return nil
}

func (s *testSources) UpdateSource(ctx context.Context, source *sources.Config) error {
	if source == nil {
		return sources.ErrInvalidSource
	}
	if err := s.ValidateSource(source); err != nil {
		return err
	}

	// Set default index names if empty
	if source.ArticleIndex == "" {
		source.ArticleIndex = DefaultArticleIndex
	}
	if source.Index == "" {
		source.Index = DefaultContentIndex
	}

	for i, config := range s.configs {
		if config.Name == source.Name {
			s.configs[i] = *source
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

func (s *testSources) DeleteSource(ctx context.Context, name string) error {
	for i, config := range s.configs {
		if config.Name == name {
			s.configs = append(s.configs[:i], s.configs[i+1:]...)
			return nil
		}
	}
	return sources.ErrSourceNotFound
}

func (s *testSources) ValidateSource(source *sources.Config) error {
	if source == nil {
		return sources.ErrInvalidSource
	}
	if source.Name == "" {
		return errors.New(ErrInvalidSourceName)
	}
	if source.URL == "" {
		return errors.New(ErrInvalidSourceURL)
	}
	if source.RateLimit <= 0 {
		return errors.New(ErrInvalidRateLimit)
	}
	if source.MaxDepth <= 0 {
		return errors.New(ErrInvalidMaxDepth)
	}
	return nil
}

func (s *testSources) GetMetrics() sources.Metrics {
	return sources.Metrics{
		SourceCount: int64(len(s.configs)),
	}
}

// GetSources retrieves all source configurations.
func (s *testSources) GetSources() ([]sources.Config, error) {
	return s.configs, nil
}

// FindByName finds a source by name.
func (s *testSources) FindByName(name string) (*sources.Config, error) {
	for _, config := range s.configs {
		if config.Name == name {
			return &config, nil
		}
	}
	return nil, sources.ErrSourceNotFound
}

func (s *testSources) SetSources(configs []sources.Config) {
	s.configs = configs
}
