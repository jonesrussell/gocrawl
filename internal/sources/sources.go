// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"

	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Config represents a source configuration.
type Config = sourceutils.SourceConfig

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig = sourceutils.SelectorConfig

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors = sourceutils.ArticleSelectors

// Sources manages a collection of web content sources.
type Sources struct {
	sources []Config
	logger  logger.Interface
	metrics *sourceutils.SourcesMetrics
}

// Ensure Sources implements Interface
var _ Interface = (*Sources)(nil)

// createSelectorConfig creates a new SelectorConfig from the given selectors
func createSelectorConfig(selectors any) sourceutils.SelectorConfig {
	var articleSelectors sourceutils.ArticleSelectors
	var listSelectors sourceutils.ListSelectors

	switch s := selectors.(type) {
	case configtypes.SourceSelectors:
		articleSelectors = sourceutils.ArticleSelectors{
			Container:     s.Article.Container,
			Title:         s.Article.Title,
			Body:          s.Article.Body,
			Intro:         s.Article.Intro,
			Link:          s.Article.Link,
			Image:         s.Article.Image,
			Byline:        s.Article.Byline,
			PublishedTime: s.Article.PublishedTime,
			TimeAgo:       s.Article.TimeAgo,
			JSONLD:        s.Article.JSONLD,
			Section:       s.Article.Section,
			Keywords:      s.Article.Keywords,
			Description:   s.Article.Description,
			OGTitle:       s.Article.OGTitle,
			OGDescription: s.Article.OGDescription,
			OGImage:       s.Article.OGImage,
			OGType:        s.Article.OGType,
			OGSiteName:    s.Article.OGSiteName,
			OgURL:         s.Article.OgURL,
			Canonical:     s.Article.Canonical,
			WordCount:     s.Article.WordCount,
			PublishDate:   s.Article.PublishDate,
			Category:      s.Article.Category,
			Tags:          s.Article.Tags,
			Author:        s.Article.Author,
			BylineName:    s.Article.BylineName,
			ArticleID:     s.Article.ArticleID,
			Exclude:       s.Article.Exclude,
		}
		listSelectors = sourceutils.ListSelectors{
			Container:       s.List.Container,
			ArticleCards:    s.List.ArticleCards,
			ArticleList:     s.List.ArticleList,
			ExcludeFromList: s.List.ExcludeFromList,
		}
	case configtypes.ArticleSelectors:
		articleSelectors = sourceutils.ArticleSelectors{
			Container:     s.Container,
			Title:         s.Title,
			Body:          s.Body,
			Intro:         s.Intro,
			Link:          s.Link,
			Image:         s.Image,
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
			OGType:        s.OGType,
			OGSiteName:    s.OGSiteName,
			OgURL:         s.OgURL,
			Canonical:     s.Canonical,
			WordCount:     s.WordCount,
			PublishDate:   s.PublishDate,
			Category:      s.Category,
			Tags:          s.Tags,
			Author:        s.Author,
			BylineName:    s.BylineName,
			ArticleID:     s.ArticleID,
			Exclude:       s.Exclude,
		}
	case loader.SourceSelectors:
		articleSelectors = sourceutils.ArticleSelectors{
			Container:     s.Article.Container,
			Title:         s.Article.Title,
			Body:          s.Article.Body,
			Intro:         s.Article.Intro,
			Link:          s.Article.Link,
			Image:         s.Article.Image,
			Byline:        s.Article.Byline,
			PublishedTime: s.Article.PublishedTime,
			TimeAgo:       s.Article.TimeAgo,
			JSONLD:        s.Article.JSONLD,
			Section:       s.Article.Section,
			Keywords:      s.Article.Keywords,
			Description:   s.Article.Description,
			OGTitle:       s.Article.OGTitle,
			OGDescription: s.Article.OGDescription,
			OGImage:       s.Article.OGImage,
			OGType:        s.Article.OGType,
			OGSiteName:    s.Article.OGSiteName,
			OgURL:         s.Article.OgURL,
			Canonical:     s.Article.Canonical,
			WordCount:     s.Article.WordCount,
			PublishDate:   s.Article.PublishDate,
			Category:      s.Article.Category,
			Tags:          s.Article.Tags,
			Author:        s.Article.Author,
			BylineName:    s.Article.BylineName,
			ArticleID:     s.Article.ArticleID,
			Exclude:       s.Article.Exclude,
		}
		listSelectors = sourceutils.ListSelectors{
			Container:       s.List.Container,
			ArticleCards:    s.List.ArticleCards,
			ArticleList:     s.List.ArticleList,
			ExcludeFromList: s.List.ExcludeFromList,
		}
	case loader.ArticleSelectors:
		articleSelectors = sourceutils.ArticleSelectors{
			Container:     s.Container,
			Title:         s.Title,
			Body:          s.Body,
			Intro:         s.Intro,
			Link:          s.Link,
			Image:         s.Image,
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
			OGType:        s.OGType,
			OGSiteName:    s.OGSiteName,
			OgURL:         s.OgURL,
			Canonical:     s.Canonical,
			WordCount:     s.WordCount,
			PublishDate:   s.PublishDate,
			Category:      s.Category,
			Tags:          s.Tags,
			Author:        s.Author,
			BylineName:    s.BylineName,
			ArticleID:     s.ArticleID,
			Exclude:       s.Exclude,
		}
	default:
		// Return empty selectors for unknown types
		articleSelectors = sourceutils.ArticleSelectors{}
		listSelectors = sourceutils.ListSelectors{}
	}

	return sourceutils.SelectorConfig{
		Article: articleSelectors,
		List:    listSelectors,
	}
}

// LoadSources creates a new Sources instance by loading sources from either:
// 1. A YAML file specified in the crawler config
// Returns an error if no sources are found
// The logger parameter is optional and can be nil, but it's recommended to provide one
// for consistency with LoadFromFile and to avoid potential nil pointer dereferences.
func LoadSources(cfg config.Interface, logger logger.Interface) (*Sources, error) {
	// Determine source file path, handling nil crawler config
	sourceFile := "sources.yml" // Default source file
	if crawlerCfg := cfg.GetCrawlerConfig(); crawlerCfg != nil {
		if crawlerCfg.SourceFile != "" {
			sourceFile = crawlerCfg.SourceFile
		}
	}

	// Load sources from config file
	sources, err := loadSourcesFromFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources from file: %w", err)
	}

	// If no sources found, return an error
	if len(sources) == 0 {
		return nil, errors.New("no sources found in file")
	}

	return &Sources{
		sources: sources,
		logger:  logger, // Initialize logger field for consistency with LoadFromFile
		metrics: sourceutils.NewSourcesMetrics(),
	}, nil
}

// loadSourcesFromFile attempts to load sources from a file
func loadSourcesFromFile(path string) ([]Config, error) {
	sourceLoader, err := loader.NewLoader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create source loader: %w", err)
	}

	configs, err := sourceLoader.LoadSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load sources: %w", err)
	}

	if len(configs) == 0 {
		return nil, errors.New("no sources found in file")
	}

	// Convert loaded configs to our source type
	sourceConfigs := make([]Config, 0, len(configs))
	for i := range configs {
		sourceConfigs = append(sourceConfigs, convertLoaderConfig(configs[i]))
	}

	return sourceConfigs, nil
}

// convertLoaderConfig converts a loader.Config to a sourceutils.SourceConfig
func convertLoaderConfig(cfg loader.Config) Config {
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
			switch val := v.(type) {
			case int:
				rateLimit = time.Duration(val) * time.Second
			case int64:
				rateLimit = time.Duration(val) * time.Second
			case float64:
				rateLimit = time.Duration(val) * time.Second
			default:
				rateLimit = time.Second
			}
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
		return Config{
			Name:           cfg.Name,
			URL:            cfg.URL,
			AllowedDomains: []string{cfg.URL},
			StartURLs:      []string{cfg.URL},
			RateLimit:      rateLimit,
			MaxDepth:       cfg.MaxDepth,
			Time:           cfg.Time,
			Index:          cfg.Index,
			ArticleIndex:   cfg.ArticleIndex,
			Selectors:      createSelectorConfig(cfg.Selectors),
			Rules:          configtypes.Rules{},
		}
	}

	// Get the domain from the URL
	domain := u.Hostname()
	if domain == "" {
		domain = cfg.URL
	}

	return Config{
		Name:           cfg.Name,
		URL:            cfg.URL,
		AllowedDomains: []string{domain},
		StartURLs:      []string{cfg.URL},
		RateLimit:      rateLimit,
		MaxDepth:       cfg.MaxDepth,
		Time:           cfg.Time,
		Index:          cfg.Index,
		ArticleIndex:   cfg.ArticleIndex,
		Selectors:      createSelectorConfig(cfg.Selectors),
		Rules:          configtypes.Rules{},
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
	sourceConfigs := make([]Config, 0, len(configs))
	for i := range configs {
		sourceConfigs = append(sourceConfigs, convertLoaderConfig(configs[i]))
	}

	sources.sources = sourceConfigs
	return sources, nil
}

// SetSources sets the sources.
func (s *Sources) SetSources(configs []Config) {
	s.sources = configs
	s.metrics.SourceCount = int64(len(configs))
	s.metrics.LastUpdated = time.Now()
}

// ListSources retrieves all sources.
func (s *Sources) ListSources(ctx context.Context) ([]*Config, error) {
	result := make([]*Config, 0, len(s.sources))
	for i := range s.sources {
		result = append(result, &s.sources[i])
	}
	return result, nil
}

// AddSource adds a new source.
func (s *Sources) AddSource(ctx context.Context, source *Config) error {
	// Validate the source configuration
	if source == nil {
		return ErrInvalidSource
	}

	// Check if source already exists
	if s.FindByName(source.Name) != nil {
		return ErrSourceExists
	}

	s.sources = append(s.sources, *source)
	s.metrics.SourceCount = int64(len(s.sources))
	s.metrics.LastUpdated = time.Now()
	return nil
}

// UpdateSource updates an existing source.
func (s *Sources) UpdateSource(ctx context.Context, source *Config) error {
	// Validate the source configuration
	if source == nil {
		return ErrInvalidSource
	}

	// Check if source exists
	if s.FindByName(source.Name) == nil {
		return ErrSourceNotFound
	}

	for i := range s.sources {
		if s.sources[i].Name == source.Name {
			s.sources[i] = *source
			s.metrics.LastUpdated = time.Now()
			return nil
		}
	}
	return ErrSourceNotFound
}

// DeleteSource deletes a source by name.
func (s *Sources) DeleteSource(ctx context.Context, name string) error {
	for i := range s.sources {
		if s.sources[i].Name == name {
			s.sources = append(s.sources[:i], s.sources[i+1:]...)
			s.metrics.SourceCount = int64(len(s.sources))
			s.metrics.LastUpdated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("source not found: %s", name)
}

// ValidateSource validates a source configuration and returns the validated source.
// It checks if the source exists and is properly configured.
func (s *Sources) ValidateSource(
	ctx context.Context,
	sourceName string,
	indexManager storagetypes.IndexManager,
) (*configtypes.Source, error) {
	// Get all sources
	sourceConfigs, err := s.GetSources()
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}

	// If no sources are configured, return an error
	if len(sourceConfigs) == 0 {
		return nil, errors.New("no sources configured")
	}

	// Find the requested source
	var selectedSource *Config
	for i := range sourceConfigs {
		if sourceConfigs[i].Name == sourceName {
			selectedSource = &sourceConfigs[i]
			break
		}
	}

	// If source not found, return an error
	if selectedSource == nil {
		return nil, fmt.Errorf("source not found: %s", sourceName)
	}

	// Convert to configtypes.Source
	source := sourceutils.ConvertToConfigSource(selectedSource)

	// Ensure article index exists if specified
	if selectedSource.ArticleIndex != "" {
		if indexErr := indexManager.EnsureArticleIndex(ctx, selectedSource.ArticleIndex); indexErr != nil {
			return nil, fmt.Errorf("failed to ensure article index exists: %w", indexErr)
		}
	}

	// Ensure page index exists if specified
	if selectedSource.Index != "" {
		if pageErr := indexManager.EnsurePageIndex(ctx, selectedSource.Index); pageErr != nil {
			return nil, fmt.Errorf("failed to ensure page index exists: %w", pageErr)
		}
	}

	return source, nil
}

// GetMetrics returns the current metrics.
func (s *Sources) GetMetrics() sourceutils.SourcesMetrics {
	return *s.metrics
}

// GetSources returns all sources.
func (s *Sources) GetSources() ([]Config, error) {
	return s.sources, nil
}

// FindByName finds a source by name.
func (s *Sources) FindByName(name string) *Config {
	for i := range s.sources {
		if s.sources[i].Name == name {
			return &s.sources[i]
		}
	}
	return nil
}

// articleSelector represents the selectors for article content.
type articleSelector struct {
	Container     string   `yaml:"container"`
	Title         string   `yaml:"title"`
	Body          string   `yaml:"body"`
	Intro         string   `yaml:"intro"`
	Link          string   `yaml:"link"`
	Image         string   `yaml:"image"`
	Byline        string   `yaml:"byline"`
	PublishedTime string   `yaml:"published_time"`
	TimeAgo       string   `yaml:"time_ago"`
	JSONLD        string   `yaml:"jsonld"`
	Section       string   `yaml:"section"`
	Keywords      string   `yaml:"keywords"`
	Description   string   `yaml:"description"`
	OGTitle       string   `yaml:"og_title"`
	OGDescription string   `yaml:"og_description"`
	OGImage       string   `yaml:"og_image"`
	OGType        string   `yaml:"og_type"`
	OGSiteName    string   `yaml:"og_site_name"`
	OgURL         string   `yaml:"og_url"`
	Canonical     string   `yaml:"canonical"`
	WordCount     string   `yaml:"word_count"`
	PublishDate   string   `yaml:"publish_date"`
	Category      string   `yaml:"category"`
	Tags          string   `yaml:"tags"`
	Author        string   `yaml:"author"`
	BylineName    string   `yaml:"byline_name"`
	ArticleID     string   `yaml:"article_id"`
	Exclude       []string `yaml:"exclude"`
}

// getArticleSelectorsFromSelector converts an articleSelector to ArticleSelectors.
func getArticleSelectorsFromSelector(s articleSelector) ArticleSelectors {
	return ArticleSelectors{
		Container:     s.Container,
		Title:         s.Title,
		Body:          s.Body,
		Intro:         s.Intro,
		Link:          s.Link,
		Image:         s.Image,
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
		OGType:        s.OGType,
		OGSiteName:    s.OGSiteName,
		OgURL:         s.OgURL,
		Canonical:     s.Canonical,
		WordCount:     s.WordCount,
		PublishDate:   s.PublishDate,
		Category:      s.Category,
		Tags:          s.Tags,
		Author:        s.Author,
		BylineName:    s.BylineName,
		ArticleID:     s.ArticleID,
		Exclude:       s.Exclude,
	}
}

// extractArticleSelectorsFromLoader converts loader.ArticleSelectors to articleSelector.
func extractArticleSelectorsFromLoader(selectors loader.ArticleSelectors) articleSelector {
	return articleSelector{
		Container:     selectors.Container,
		Title:         selectors.Title,
		Body:          selectors.Body,
		Intro:         selectors.Intro,
		Link:          selectors.Link,
		Image:         selectors.Image,
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
		OGType:        selectors.OGType,
		OGSiteName:    selectors.OGSiteName,
		OgURL:         selectors.OgURL,
		Canonical:     selectors.Canonical,
		WordCount:     selectors.WordCount,
		PublishDate:   selectors.PublishDate,
		Category:      selectors.Category,
		Tags:          selectors.Tags,
		Author:        selectors.Author,
		BylineName:    selectors.BylineName,
		ArticleID:     selectors.ArticleID,
		Exclude:       selectors.Exclude,
	}
}

// extractArticleSelectorsFromConfig converts configtypes.ArticleSelectors to articleSelector.
func extractArticleSelectorsFromConfig(selectors configtypes.ArticleSelectors) articleSelector {
	return articleSelector{
		Container:     selectors.Container,
		Title:         selectors.Title,
		Body:          selectors.Body,
		Intro:         selectors.Intro,
		Link:          selectors.Link,
		Image:         selectors.Image,
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
		OGType:        selectors.OGType,
		OGSiteName:    selectors.OGSiteName,
		OgURL:         selectors.OgURL,
		Canonical:     selectors.Canonical,
		WordCount:     selectors.WordCount,
		PublishDate:   selectors.PublishDate,
		Category:      selectors.Category,
		Tags:          selectors.Tags,
		Author:        selectors.Author,
		BylineName:    selectors.BylineName,
		ArticleID:     selectors.ArticleID,
		Exclude:       selectors.Exclude,
	}
}

// NewSelectorConfigFromLoader creates a new SelectorConfig from a loader.Config.
func NewSelectorConfigFromLoader(src loader.Config) SelectorConfig {
	return SelectorConfig{
		Article: getArticleSelectorsFromSelector(extractArticleSelectorsFromLoader(src.Selectors.Article)),
		List: sourceutils.ListSelectors{
			Container:       src.Selectors.List.Container,
			ArticleCards:    src.Selectors.List.ArticleCards,
			ArticleList:     src.Selectors.List.ArticleList,
			ExcludeFromList: src.Selectors.List.ExcludeFromList,
		},
	}
}

// NewSelectorConfigFromSource creates a new SelectorConfig from a configtypes.Source.
func NewSelectorConfigFromSource(src configtypes.Source) SelectorConfig {
	return SelectorConfig{
		Article: getArticleSelectorsFromSelector(extractArticleSelectorsFromConfig(src.Selectors.Article)),
		List: sourceutils.ListSelectors{
			Container:       src.Selectors.List.Container,
			ArticleCards:    src.Selectors.List.ArticleCards,
			ArticleList:     src.Selectors.List.ArticleList,
			ExcludeFromList: src.Selectors.List.ExcludeFromList,
		},
	}
}
