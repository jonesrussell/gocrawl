// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
)

// Config represents a source configuration.
type Config struct {
	Name         string            `yaml:"name"`
	URL          string            `yaml:"url"`
	RateLimit    time.Duration     `yaml:"-"`
	MaxDepth     int               `yaml:"max_depth"`
	ArticleIndex string            `yaml:"article_index"`
	Index        string            `yaml:"index"`
	Time         []string          `yaml:"time,omitempty"`
	Selectors    SelectorConfig    `yaml:"selectors"`
	Metadata     map[string]string `yaml:"metadata,omitempty"`
}

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors struct {
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

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig struct {
	Article ArticleSelectors `yaml:"article"`
}

// Sources manages web content source configurations.
type Sources struct {
	sources []Config
	logger  interface {
		Debug(msg string, fields ...any)
		Info(msg string, fields ...any)
		Warn(msg string, fields ...any)
		Error(msg string, fields ...any)
	}
	metrics Metrics
}

// LoadFromFile loads source configurations from a YAML file.
// It reads and parses the file, populating the Sources struct with the configuration data.
//
// Parameters:
//   - path: Path to the YAML configuration file
//
// Returns:
//   - *Sources: The loaded sources configuration
//   - error: Any error that occurred during loading
func LoadFromFile(path string, logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}) (*Sources, error) {
	loaderConfigs, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	var configs []Config
	for _, src := range loaderConfigs {
		rateLimit, parseErr := time.ParseDuration(src.RateLimit)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse rate limit for source %s: %w", src.Name, parseErr)
		}

		configs = append(configs, Config{
			Name:      src.Name,
			URL:       src.URL,
			RateLimit: rateLimit,
			MaxDepth:  src.MaxDepth,
			Time:      src.Time,
		})
	}

	return &Sources{
		sources: configs,
		logger:  logger,
		metrics: Metrics{
			SourceCount: int64(len(configs)),
		},
	}, nil
}

// SetSources sets the sources for testing purposes.
func (s *Sources) SetSources(configs []Config) {
	s.sources = configs
	s.metrics.SourceCount = int64(len(configs))
}

// GetSource retrieves a source by name.
func (s *Sources) GetSource(ctx context.Context, name string) (*Config, error) {
	for _, src := range s.sources {
		if src.Name == name {
			return &src, nil
		}
	}
	return nil, ErrSourceNotFound
}

// ListSources retrieves all sources.
func (s *Sources) ListSources(ctx context.Context) ([]*Config, error) {
	result := make([]*Config, len(s.sources))
	for i := range s.sources {
		result[i] = &s.sources[i]
	}
	return result, nil
}

// AddSource adds a new source.
func (s *Sources) AddSource(ctx context.Context, source *Config) error {
	if err := s.ValidateSource(source); err != nil {
		return err
	}

	for _, existing := range s.sources {
		if existing.Name == source.Name {
			return ErrSourceExists
		}
	}

	s.sources = append(s.sources, *source)
	s.metrics.SourceCount = int64(len(s.sources))
	s.metrics.LastUpdated = time.Now()
	return nil
}

// UpdateSource updates an existing source.
func (s *Sources) UpdateSource(ctx context.Context, source *Config) error {
	if err := s.ValidateSource(source); err != nil {
		return err
	}

	for i, existing := range s.sources {
		if existing.Name == source.Name {
			s.sources[i] = *source
			s.metrics.LastUpdated = time.Now()
			return nil
		}
	}

	return ErrSourceNotFound
}

// DeleteSource deletes a source by name.
func (s *Sources) DeleteSource(ctx context.Context, name string) error {
	for i, src := range s.sources {
		if src.Name == name {
			s.sources = append(s.sources[:i], s.sources[i+1:]...)
			s.metrics.SourceCount = int64(len(s.sources))
			s.metrics.LastUpdated = time.Now()
			return nil
		}
	}
	return ErrSourceNotFound
}

// ValidateSource validates a source configuration.
func (s *Sources) ValidateSource(source *Config) error {
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
func (s *Sources) GetSources() ([]Config, error) {
	return s.sources, nil
}

// FindByName finds a source by name.
func (s *Sources) FindByName(name string) (*Config, error) {
	for _, src := range s.sources {
		if src.Name == name {
			return &src, nil
		}
	}
	return nil, ErrSourceNotFound
}

// newArticleSelectors creates ArticleSelectors from a source with selectors
type selectorSource interface {
	GetArticleSelectors() ArticleSelectors
}

// loaderConfigWrapper wraps loader.Config to implement selectorSource
type loaderConfigWrapper struct {
	loader.Config
}

func (w loaderConfigWrapper) GetArticleSelectors() ArticleSelectors {
	// Return empty selectors since we no longer have selectors in the Source struct
	return ArticleSelectors{}
}

// sourceWrapper wraps config.Source to implement selectorSource
type sourceWrapper struct {
	config.Source
}

func (w sourceWrapper) GetArticleSelectors() ArticleSelectors {
	// Return empty selectors since we no longer have selectors in the Source struct
	return ArticleSelectors{}
}

// newArticleSelectors creates ArticleSelectors from a source with selectors
func newArticleSelectors(src selectorSource) ArticleSelectors {
	return src.GetArticleSelectors()
}

// NewSelectorConfigFromLoader creates a new SelectorConfig from a loader source
func NewSelectorConfigFromLoader(src loader.Config) SelectorConfig {
	return SelectorConfig{
		Article: newArticleSelectors(loaderConfigWrapper{src}),
	}
}

// NewSelectorConfigFromSource creates a new SelectorConfig from a config source
func NewSelectorConfigFromSource(src config.Source) SelectorConfig {
	return SelectorConfig{
		Article: newArticleSelectors(sourceWrapper{src}),
	}
}
