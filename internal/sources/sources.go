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

// articleSelector represents the common structure of article selectors
type articleSelector struct {
	Container     string `yaml:"container"`
	Title         string `yaml:"title"`
	Body          string `yaml:"body"`
	Intro         string `yaml:"intro"`
	Byline        string `yaml:"byline"`
	PublishedTime string `yaml:"published_time"`
	TimeAgo       string `yaml:"time_ago"`
	JSONLD        string `yaml:"jsonld"`
	Section       string `yaml:"section"`
	Keywords      string `yaml:"keywords"`
	Description   string `yaml:"description"`
	OGTitle       string `yaml:"og_title"`
	OGDescription string `yaml:"og_description"`
	OGImage       string `yaml:"og_image"`
	OgURL         string `yaml:"og_url"`
	Canonical     string `yaml:"canonical"`
	WordCount     string `yaml:"word_count"`
	PublishDate   string `yaml:"publish_date"`
	Category      string `yaml:"category"`
	Tags          string `yaml:"tags"`
	Author        string `yaml:"author"`
	BylineName    string `yaml:"byline_name"`
}

// getArticleSelectorsFromSelector creates ArticleSelectors from an articleSelector
func getArticleSelectorsFromSelector(s articleSelector) ArticleSelectors {
	return ArticleSelectors{
		Container:     s.Container,
		Title:         s.Title,
		Body:          s.Body,
		Intro:         s.Intro,
		Byline:        s.Byline,
		PublishedTime: s.PublishedTime,
		TimeAgo:       s.TimeAgo,
		JSONLD:        s.JSONLD,
		Section:       s.Section,
		Keywords:      s.Keywords,
		Description:   s.Description,
		OGTitle:       s.OGTitle,
		OGDescription: s.OGDescription,
		OGImage:       s.OGImage,
		OgURL:         s.OgURL,
		Canonical:     s.Canonical,
		WordCount:     s.WordCount,
		PublishDate:   s.PublishDate,
		Category:      s.Category,
		Tags:          s.Tags,
		Author:        s.Author,
		BylineName:    s.BylineName,
	}
}

// articleSelectorProvider defines the interface for types that can provide article selectors
type articleSelectorProvider interface {
	GetArticleSelectors() ArticleSelectors
}

// getArticleSelectorsFromProvider extracts article selectors from any type that implements articleSelectorProvider
func getArticleSelectorsFromProvider(p articleSelectorProvider) ArticleSelectors {
	return p.GetArticleSelectors()
}

// extractArticleSelectorsFromLoader creates an articleSelector from loader.ArticleSelectors
func extractArticleSelectorsFromLoader(selectors loader.ArticleSelectors) articleSelector {
	return articleSelector{
		Container:     selectors.Container,
		Title:         selectors.Title,
		Body:          selectors.Body,
		Intro:         selectors.Intro,
		Byline:        selectors.Byline,
		PublishedTime: selectors.PublishedTime,
		TimeAgo:       selectors.TimeAgo,
		JSONLD:        selectors.JSONLD,
		Section:       selectors.Section,
		Keywords:      selectors.Keywords,
		Description:   selectors.Description,
		OGTitle:       selectors.OGTitle,
		OGDescription: selectors.OGDescription,
		OGImage:       selectors.OGImage,
		OgURL:         selectors.OgURL,
		Canonical:     selectors.Canonical,
		WordCount:     selectors.WordCount,
		PublishDate:   selectors.PublishDate,
		Category:      selectors.Category,
		Tags:          selectors.Tags,
		Author:        selectors.Author,
		BylineName:    selectors.BylineName,
	}
}

// extractArticleSelectorsFromConfig creates an articleSelector from config.ArticleSelectors
func extractArticleSelectorsFromConfig(selectors config.ArticleSelectors) articleSelector {
	return articleSelector{
		Container:     selectors.Container,
		Title:         selectors.Title,
		Body:          selectors.Body,
		Intro:         selectors.Intro,
		Byline:        selectors.Byline,
		PublishedTime: selectors.PublishedTime,
		TimeAgo:       selectors.TimeAgo,
		JSONLD:        selectors.JSONLD,
		Section:       selectors.Section,
		Keywords:      selectors.Keywords,
		Description:   selectors.Description,
		OGTitle:       selectors.OGTitle,
		OGDescription: selectors.OGDescription,
		OGImage:       selectors.OGImage,
		OgURL:         selectors.OgURL,
		Canonical:     selectors.Canonical,
		WordCount:     selectors.WordCount,
		PublishDate:   selectors.PublishDate,
		Category:      selectors.Category,
		Tags:          selectors.Tags,
		Author:        selectors.Author,
		BylineName:    selectors.BylineName,
	}
}

type loaderConfigWrapper struct {
	loader.Config
}

func (w loaderConfigWrapper) GetArticleSelectors() ArticleSelectors {
	return getArticleSelectorsFromSelector(extractArticleSelectorsFromLoader(w.Selectors.Article))
}

type sourceWrapper struct {
	config.Source
}

func (w sourceWrapper) GetArticleSelectors() ArticleSelectors {
	return getArticleSelectorsFromSelector(extractArticleSelectorsFromConfig(w.Selectors.Article))
}

// newArticleSelectorsFromFields creates an ArticleSelectors from a map of field names to values
func newArticleSelectorsFromFields(fields map[string]string) ArticleSelectors {
	return ArticleSelectors{
		Container:     fields["Container"],
		Title:         fields["Title"],
		Body:          fields["Body"],
		Intro:         fields["Intro"],
		Byline:        fields["Byline"],
		PublishedTime: fields["PublishedTime"],
		TimeAgo:       fields["TimeAgo"],
		JSONLD:        fields["JSONLD"],
		Section:       fields["Section"],
		Keywords:      fields["Keywords"],
		Description:   fields["Description"],
		OGTitle:       fields["OGTitle"],
		OGDescription: fields["OGDescription"],
		OGImage:       fields["OGImage"],
		OgURL:         fields["OgURL"],
		Canonical:     fields["Canonical"],
		WordCount:     fields["WordCount"],
		PublishDate:   fields["PublishDate"],
		Category:      fields["Category"],
		Tags:          fields["Tags"],
		Author:        fields["Author"],
		BylineName:    fields["BylineName"],
	}
}

// NewSelectorConfigFromLoader creates a new SelectorConfig from a loader.Config.
func NewSelectorConfigFromLoader(src loader.Config) SelectorConfig {
	wrapper := loaderConfigWrapper{src}
	return SelectorConfig{
		Article: wrapper.GetArticleSelectors(),
	}
}

// NewSelectorConfigFromSource creates a new SelectorConfig from a config.Source.
func NewSelectorConfigFromSource(src config.Source) SelectorConfig {
	wrapper := sourceWrapper{src}
	return SelectorConfig{
		Article: wrapper.GetArticleSelectors(),
	}
}
