// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/types"
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
	metrics *sourceutils.SourcesMetrics
}

// Ensure Sources implements Interface
var _ Interface = (*Sources)(nil)

// ConvertSourceConfig converts a sources.Config to a types.Source.
// It handles the conversion of fields between the two types.
func ConvertSourceConfig(source *Config) *types.Source {
	if source == nil {
		return nil
	}

	return sourceutils.ConvertToConfigSource(source)
}

// convertSourceConfig converts a types.Source to a sourceutils.SourceConfig
func convertSourceConfig(src types.Source) sourceutils.SourceConfig {
	// Parse the rate limit duration
	rateLimit, err := time.ParseDuration(src.RateLimit)
	if err != nil {
		// If parsing fails, use a default value
		rateLimit = time.Second
	}

	return sourceutils.SourceConfig{
		Name:           src.Name,
		URL:            src.URL,
		AllowedDomains: src.AllowedDomains,
		StartURLs:      src.StartURLs,
		RateLimit:      rateLimit,
		MaxDepth:       src.MaxDepth,
		Time:           src.Time,
		Index:          src.Index,
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
		Rules: src.Rules,
	}
}

// LoadSources creates a new Sources instance by loading sources from either:
// 1. A YAML file specified in the crawler config
// 2. The configuration object itself
// 3. A default source based on the command line argument
// Returns an error if no sources are found in any location
func LoadSources(cfg config.Interface) (*Sources, error) {
	sources := &Sources{
		metrics: sourceutils.NewSourcesMetrics(),
	}

	// Try to load sources from file first
	crawlerCfg := cfg.GetCrawlerConfig()
	if crawlerCfg != nil && crawlerCfg.SourceFile != "" {
		if configs, err := loadSourcesFromFile(crawlerCfg.SourceFile); err == nil && len(configs) > 0 {
			sources.SetSources(configs)
			return sources, nil
		}
	}

	// Fall back to sources from config
	if srcs := cfg.GetSources(); len(srcs) > 0 {
		configs := make([]sourceutils.SourceConfig, 0, len(srcs))
		for i := range srcs {
			configs = append(configs, convertSourceConfig(srcs[i]))
		}
		sources.SetSources(configs)
		return sources, nil
	}

	// If we get here, we couldn't find any sources in the config
	// Try to create a default source based on the command line argument
	if cmd := cfg.GetCommand(); cmd != "" {
		// Create a default source based on the command
		defaultSource := &types.Source{
			Name:           cmd,
			URL:            fmt.Sprintf("https://%s", strings.ReplaceAll(cmd, " ", "")),
			AllowedDomains: []string{strings.ReplaceAll(cmd, " ", "")},
			StartURLs:      []string{fmt.Sprintf("https://%s", strings.ReplaceAll(cmd, " ", ""))},
			MaxDepth:       DefaultMaxDepth,
			RateLimit:      DefaultRateLimit.String(),
			Index:          "content",
			Rules:          types.Rules{},
		}

		// Convert to sourceutils.SourceConfig
		sourceConfig := convertSourceConfig(*defaultSource)
		sources.SetSources([]sourceutils.SourceConfig{sourceConfig})
		return sources, nil
	}

	// No sources found
	return nil, fmt.Errorf("no sources found in configuration")
}

// loadSourcesFromFile attempts to load sources from a file
func loadSourcesFromFile(path string) ([]sourceutils.SourceConfig, error) {
	sourceLoader, err := loader.NewLoader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create source loader: %w", err)
	}

	configs, err := sourceLoader.LoadSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no sources found in file")
	}

	// Convert loaded configs to our source type
	sourceConfigs := make([]sourceutils.SourceConfig, 0, len(configs))
	for i := range configs {
		sourceConfigs = append(sourceConfigs, convertLoaderConfig(configs[i]))
	}

	return sourceConfigs, nil
}

// convertLoaderConfig converts a loader.Config to a sourceutils.SourceConfig
func convertLoaderConfig(cfg loader.Config) sourceutils.SourceConfig {
	// Parse rate limit duration
	var rateLimit time.Duration
	if cfg.RateLimit != nil {
		switch v := cfg.RateLimit.(type) {
		case string:
			var err error
			rateLimit, err = time.ParseDuration(v)
			if err != nil {
				// If parsing fails, use a default value
				rateLimit = time.Second
			}
		case int, int64, float64:
			// Convert numeric value to duration in seconds
			rateLimit = time.Duration(v.(int)) * time.Second
		default:
			// Use default value for unknown types
			rateLimit = time.Second
		}
	} else {
		// Default to 1 second if not specified
		rateLimit = time.Second
	}

	// Parse URL to get domain
	u, err := url.Parse(cfg.URL)
	if err != nil {
		// If URL parsing fails, use the URL as is
		return sourceutils.SourceConfig{
			Name:           cfg.Name,
			URL:            cfg.URL,
			AllowedDomains: []string{cfg.URL},
			StartURLs:      []string{cfg.URL},
			RateLimit:      rateLimit,
			MaxDepth:       cfg.MaxDepth,
			Time:           cfg.Time,
			Index:          cfg.Index,
			ArticleIndex:   cfg.ArticleIndex,
			Selectors: sourceutils.SelectorConfig{
				Article: sourceutils.ArticleSelectors{
					Container:     cfg.Selectors.Article.Container,
					Title:         cfg.Selectors.Article.Title,
					Body:          cfg.Selectors.Article.Body,
					Intro:         cfg.Selectors.Article.Intro,
					Byline:        cfg.Selectors.Article.Byline,
					PublishedTime: cfg.Selectors.Article.PublishedTime,
					TimeAgo:       cfg.Selectors.Article.TimeAgo,
					JSONLD:        cfg.Selectors.Article.JSONLD,
					Section:       cfg.Selectors.Article.Section,
					Keywords:      cfg.Selectors.Article.Keywords,
					Description:   cfg.Selectors.Article.Description,
					OGTitle:       cfg.Selectors.Article.OGTitle,
					OGDescription: cfg.Selectors.Article.OGDescription,
					OGImage:       cfg.Selectors.Article.OGImage,
					OgURL:         cfg.Selectors.Article.OgURL,
					Canonical:     cfg.Selectors.Article.Canonical,
					WordCount:     cfg.Selectors.Article.WordCount,
					PublishDate:   cfg.Selectors.Article.PublishDate,
					Category:      cfg.Selectors.Article.Category,
					Tags:          cfg.Selectors.Article.Tags,
					Author:        cfg.Selectors.Article.Author,
					BylineName:    cfg.Selectors.Article.BylineName,
				},
			},
			Rules: types.Rules{},
		}
	}

	// Get the domain from the URL
	domain := u.Hostname()
	if domain == "" {
		domain = cfg.URL
	}

	return sourceutils.SourceConfig{
		Name:           cfg.Name,
		URL:            cfg.URL,
		AllowedDomains: []string{domain},
		StartURLs:      []string{cfg.URL},
		RateLimit:      rateLimit,
		MaxDepth:       cfg.MaxDepth,
		Time:           cfg.Time,
		Index:          cfg.Index,
		ArticleIndex:   cfg.ArticleIndex,
		Selectors: sourceutils.SelectorConfig{
			Article: sourceutils.ArticleSelectors{
				Container:     cfg.Selectors.Article.Container,
				Title:         cfg.Selectors.Article.Title,
				Body:          cfg.Selectors.Article.Body,
				Intro:         cfg.Selectors.Article.Intro,
				Byline:        cfg.Selectors.Article.Byline,
				PublishedTime: cfg.Selectors.Article.PublishedTime,
				TimeAgo:       cfg.Selectors.Article.TimeAgo,
				JSONLD:        cfg.Selectors.Article.JSONLD,
				Section:       cfg.Selectors.Article.Section,
				Keywords:      cfg.Selectors.Article.Keywords,
				Description:   cfg.Selectors.Article.Description,
				OGTitle:       cfg.Selectors.Article.OGTitle,
				OGDescription: cfg.Selectors.Article.OGDescription,
				OGImage:       cfg.Selectors.Article.OGImage,
				OgURL:         cfg.Selectors.Article.OgURL,
				Canonical:     cfg.Selectors.Article.Canonical,
				WordCount:     cfg.Selectors.Article.WordCount,
				PublishDate:   cfg.Selectors.Article.PublishDate,
				Category:      cfg.Selectors.Article.Category,
				Tags:          cfg.Selectors.Article.Tags,
				Author:        cfg.Selectors.Article.Author,
				BylineName:    cfg.Selectors.Article.BylineName,
			},
		},
		Rules: types.Rules{},
	}
}

// LoadFromFile loads sources from a YAML file.
func LoadFromFile(path string, logger logger.Interface) (*Sources, error) {
	sourceLoader, err := loader.NewLoader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create source loader: %w", err)
	}

	configs, err := sourceLoader.LoadSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	sources := &Sources{
		logger:  logger,
		metrics: sourceutils.NewSourcesMetrics(),
	}

	// Convert loaded configs to our source type
	sourceConfigs := make([]sourceutils.SourceConfig, 0, len(configs))
	for i := range configs {
		sourceConfigs = append(sourceConfigs, convertLoaderConfig(configs[i]))
	}

	sources.SetSources(sourceConfigs)
	return sources, nil
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
	for i := range s.sources {
		if s.sources[i].Name == source.Name {
			return ErrSourceExists
		}
	}

	// Set default index name if not provided
	if source.Index == "" {
		source.Index = "content"
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
	for i := range s.sources {
		if s.sources[i].Name == source.Name {
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
	for i := range s.sources {
		if s.sources[i].Name == name {
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

	// Convert to types.Source and validate
	typesSource := ConvertSourceConfig(source)
	return typesSource.Validate()
}

// GetMetrics returns the current metrics.
func (s *Sources) GetMetrics() sourceutils.SourcesMetrics {
	return *s.metrics
}

// GetSources returns all sources.
func (s *Sources) GetSources() ([]sourceutils.SourceConfig, error) {
	return s.sources, nil
}

// FindByName finds a source by name.
func (s *Sources) FindByName(name string) *sourceutils.SourceConfig {
	// If no sources are loaded, try to create a default source
	if len(s.sources) == 0 {
		// Create a default source based on the name
		defaultSource := &types.Source{
			Name:           name,
			URL:            fmt.Sprintf("https://%s", strings.ReplaceAll(name, " ", "")),
			AllowedDomains: []string{strings.ReplaceAll(name, " ", "")},
			StartURLs:      []string{fmt.Sprintf("https://%s", strings.ReplaceAll(name, " ", ""))},
			MaxDepth:       DefaultMaxDepth,
			RateLimit:      DefaultRateLimit.String(),
			Index:          "content",
			Rules:          types.Rules{},
		}

		// Convert to sourceutils.SourceConfig
		sourceConfig := convertSourceConfig(*defaultSource)
		s.sources = append(s.sources, sourceConfig)
		return &sourceConfig
	}

	// Search for the source in the loaded sources
	for i := range s.sources {
		if strings.EqualFold(s.sources[i].Name, name) {
			return &s.sources[i]
		}
	}

	// If not found, create a default source
	defaultSource := &types.Source{
		Name:           name,
		URL:            fmt.Sprintf("https://%s", strings.ReplaceAll(name, " ", "")),
		AllowedDomains: []string{strings.ReplaceAll(name, " ", "")},
		StartURLs:      []string{fmt.Sprintf("https://%s", strings.ReplaceAll(name, " ", ""))},
		MaxDepth:       DefaultMaxDepth,
		RateLimit:      DefaultRateLimit.String(),
		Index:          "content",
		Rules:          types.Rules{},
	}

	// Convert to sourceutils.SourceConfig
	sourceConfig := convertSourceConfig(*defaultSource)
	s.sources = append(s.sources, sourceConfig)
	return &sourceConfig
}

// articleSelector represents the selectors for article content.
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

// getArticleSelectorsFromSelector converts an articleSelector to ArticleSelectors.
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

// extractArticleSelectorsFromLoader converts loader.ArticleSelectors to articleSelector.
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

// extractArticleSelectorsFromConfig converts types.ArticleSelectors to articleSelector.
func extractArticleSelectorsFromConfig(selectors types.ArticleSelectors) articleSelector {
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

// NewSelectorConfigFromLoader creates a new SelectorConfig from a loader.Config.
func NewSelectorConfigFromLoader(src loader.Config) SelectorConfig {
	return SelectorConfig{
		Article: getArticleSelectorsFromSelector(extractArticleSelectorsFromLoader(src.Selectors.Article)),
	}
}

// NewSelectorConfigFromSource creates a new SelectorConfig from a types.Source.
func NewSelectorConfigFromSource(src types.Source) SelectorConfig {
	return SelectorConfig{
		Article: getArticleSelectorsFromSelector(extractArticleSelectorsFromConfig(src.Selectors.Article)),
	}
}
