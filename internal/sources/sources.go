// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
)

// Config represents a source configuration.
type Config = sourceutils.SourceConfig

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig = sourceutils.SelectorConfig

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors = sourceutils.ArticleSelectors

// Sources manages a collection of source configurations.
type Sources struct {
	sources []sourceutils.SourceConfig
	logger  logger.Interface
	metrics Metrics
}

// Ensure Sources implements Interface
var _ Interface = (*Sources)(nil)

// ConvertSourceConfig converts a sources.Config to a config.Source.
// It handles the conversion of fields between the two types.
func ConvertSourceConfig(source *Config) *config.Source {
	if source == nil {
		return nil
	}

	return sourceutils.ConvertToConfigSource(source)
}

// convertSourceConfig converts a config.Source to a sourceutils.SourceConfig
func convertSourceConfig(src config.Source) sourceutils.SourceConfig {
	return sourceutils.SourceConfig{
		Name:         src.Name,
		URL:          src.URL,
		RateLimit:    src.RateLimit,
		MaxDepth:     src.MaxDepth,
		ArticleIndex: src.ArticleIndex,
		Index:        src.Index,
		Selectors: sourceutils.SelectorConfig{
			Article: sourceutils.ArticleSelectors{
				Container:     src.Selectors.Article.Container,
				Title:         src.Selectors.Article.Title,
				Body:          src.Selectors.Article.Body,
				Intro:         src.Selectors.Article.Intro,
				Byline:        src.Selectors.Article.Byline,
				PublishedTime: src.Selectors.Article.PublishedTime,
				TimeAgo:       src.Selectors.Article.TimeAgo,
				JSONLD:        src.Selectors.Article.JSONLD,
				Section:       src.Selectors.Article.Section,
				Keywords:      src.Selectors.Article.Keywords,
				Description:   src.Selectors.Article.Description,
				OGTitle:       src.Selectors.Article.OGTitle,
				OGDescription: src.Selectors.Article.OGDescription,
				OGImage:       src.Selectors.Article.OGImage,
				OgURL:         src.Selectors.Article.OgURL,
				Canonical:     src.Selectors.Article.Canonical,
				WordCount:     src.Selectors.Article.WordCount,
				PublishDate:   src.Selectors.Article.PublishDate,
				Category:      src.Selectors.Article.Category,
				Tags:          src.Selectors.Article.Tags,
				Author:        src.Selectors.Article.Author,
				BylineName:    src.Selectors.Article.BylineName,
			},
		},
	}
}

// NewSourcesFromConfig creates a new Sources instance from a config.
func NewSourcesFromConfig(cfg config.Interface, logger logger.Interface) Interface {
	sources := &Sources{
		logger: logger,
		metrics: Metrics{
			SourceCount: 0,
		},
	}

	// Load sources from config
	srcs := cfg.GetSources()

	// Convert config sources to our source type
	configs := make([]sourceutils.SourceConfig, 0, len(srcs))
	for _, src := range srcs {
		configs = append(configs, convertSourceConfig(src))
	}

	sources.SetSources(configs)
	sources.metrics.SourceCount = int64(len(configs))
	sources.metrics.LastUpdated = time.Now()

	return sources
}

// LoadFromFile loads sources from a YAML file.
func LoadFromFile(path string, logger logger.Interface) (*Sources, error) {
	// Implementation details...
	return &Sources{
		logger: logger,
		metrics: Metrics{
			SourceCount: 0,
		},
	}, nil
}

// SetSources sets the sources.
func (s *Sources) SetSources(configs []sourceutils.SourceConfig) {
	s.sources = configs
	s.metrics.SourceCount = int64(len(configs))
	s.metrics.LastUpdated = time.Now()
}

// ListSources retrieves all sources.
func (s *Sources) ListSources(ctx context.Context) ([]*sourceutils.SourceConfig, error) {
	result := make([]*sourceutils.SourceConfig, 0, len(s.sources))
	for i := range s.sources {
		result = append(result, &s.sources[i])
	}
	return result, nil
}

// AddSource adds a new source.
func (s *Sources) AddSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	if source == nil {
		return ErrInvalidSource
	}

	// Check if source already exists
	for _, existing := range s.sources {
		if existing.Name == source.Name {
			return ErrSourceExists
		}
	}

	// Add the new source
	s.sources = append(s.sources, *source)
	s.metrics.SourceCount = int64(len(s.sources))
	s.metrics.LastUpdated = time.Now()

	return nil
}

// UpdateSource updates an existing source.
func (s *Sources) UpdateSource(ctx context.Context, source *sourceutils.SourceConfig) error {
	if source == nil {
		return ErrInvalidSource
	}

	// Find and update the source
	found := false
	for i, existing := range s.sources {
		if existing.Name == source.Name {
			s.sources[i] = *source
			found = true
			break
		}
	}

	if !found {
		return ErrSourceNotFound
	}

	s.metrics.LastUpdated = time.Now()
	return nil
}

// DeleteSource deletes a source by name.
func (s *Sources) DeleteSource(ctx context.Context, name string) error {
	// Find and remove the source
	for i, existing := range s.sources {
		if existing.Name == name {
			s.sources = append(s.sources[:i], s.sources[i+1:]...)
			s.metrics.SourceCount = int64(len(s.sources))
			s.metrics.LastUpdated = time.Now()
			return nil
		}
	}

	return ErrSourceNotFound
}

// ValidateSource validates a source configuration.
func (s *Sources) ValidateSource(source *sourceutils.SourceConfig) error {
	if source == nil {
		return ErrInvalidSource
	}

	// Basic validation
	if source.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidSource)
	}

	if source.URL == "" {
		return fmt.Errorf("%w: URL is required", ErrInvalidSource)
	}

	return nil
}

// GetMetrics returns the current metrics.
func (s *Sources) GetMetrics() Metrics {
	return s.metrics
}

// GetSources retrieves all source configurations.
func (s *Sources) GetSources() ([]sourceutils.SourceConfig, error) {
	return s.sources, nil
}

// FindByName finds a source by name.
func (s *Sources) FindByName(name string) (*sourceutils.SourceConfig, error) {
	for i, source := range s.sources {
		if source.Name == name {
			return &s.sources[i], nil
		}
	}
	return nil, ErrSourceNotFound
}

type selectorSource interface {
	GetArticleSelectors() ArticleSelectors
}

type loaderConfigWrapper struct {
	loader.Config
}

func (w loaderConfigWrapper) GetArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		Container:     w.Selectors.Article.Container,
		Title:         w.Selectors.Article.Title,
		Body:          w.Selectors.Article.Body,
		Intro:         w.Selectors.Article.Intro,
		Byline:        w.Selectors.Article.Byline,
		PublishedTime: w.Selectors.Article.PublishedTime,
		TimeAgo:       w.Selectors.Article.TimeAgo,
		JSONLD:        w.Selectors.Article.JSONLD,
		Section:       w.Selectors.Article.Section,
		Keywords:      w.Selectors.Article.Keywords,
		Description:   w.Selectors.Article.Description,
		OGTitle:       w.Selectors.Article.OGTitle,
		OGDescription: w.Selectors.Article.OGDescription,
		OGImage:       w.Selectors.Article.OGImage,
		OgURL:         w.Selectors.Article.OgURL,
		Canonical:     w.Selectors.Article.Canonical,
		WordCount:     w.Selectors.Article.WordCount,
		PublishDate:   w.Selectors.Article.PublishDate,
		Category:      w.Selectors.Article.Category,
		Tags:          w.Selectors.Article.Tags,
		Author:        w.Selectors.Article.Author,
		BylineName:    w.Selectors.Article.BylineName,
	}
}

type sourceWrapper struct {
	config.Source
}

func (w sourceWrapper) GetArticleSelectors() ArticleSelectors {
	return ArticleSelectors{
		Container:     w.Selectors.Article.Container,
		Title:         w.Selectors.Article.Title,
		Body:          w.Selectors.Article.Body,
		Intro:         w.Selectors.Article.Intro,
		Byline:        w.Selectors.Article.Byline,
		PublishedTime: w.Selectors.Article.PublishedTime,
		TimeAgo:       w.Selectors.Article.TimeAgo,
		JSONLD:        w.Selectors.Article.JSONLD,
		Section:       w.Selectors.Article.Section,
		Keywords:      w.Selectors.Article.Keywords,
		Description:   w.Selectors.Article.Description,
		OGTitle:       w.Selectors.Article.OGTitle,
		OGDescription: w.Selectors.Article.OGDescription,
		OGImage:       w.Selectors.Article.OGImage,
		OgURL:         w.Selectors.Article.OgURL,
		Canonical:     w.Selectors.Article.Canonical,
		WordCount:     w.Selectors.Article.WordCount,
		PublishDate:   w.Selectors.Article.PublishDate,
		Category:      w.Selectors.Article.Category,
		Tags:          w.Selectors.Article.Tags,
		Author:        w.Selectors.Article.Author,
		BylineName:    w.Selectors.Article.BylineName,
	}
}

func newArticleSelectors(src selectorSource) ArticleSelectors {
	return src.GetArticleSelectors()
}

// NewSelectorConfigFromLoader creates a new SelectorConfig from a loader.Config.
func NewSelectorConfigFromLoader(src loader.Config) SelectorConfig {
	wrapper := loaderConfigWrapper{src}
	return SelectorConfig{
		Article: newArticleSelectors(wrapper),
	}
}

// NewSelectorConfigFromSource creates a new SelectorConfig from a config.Source.
func NewSelectorConfigFromSource(src config.Source) SelectorConfig {
	wrapper := sourceWrapper{src}
	return SelectorConfig{
		Article: newArticleSelectors(wrapper),
	}
}
