// Package sources manages the configuration and lifecycle of web content sources for GoCrawl.
// It handles source configuration loading and validation through a YAML-based configuration system.
package sources

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"

	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/sources/types"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Interface defines the interface for managing sources.
type Interface interface {
	// ListSources retrieves all sources.
	ListSources(ctx context.Context) ([]*Config, error)
	// AddSource adds a new source.
	AddSource(ctx context.Context, source *Config) error
	// UpdateSource updates an existing source.
	UpdateSource(ctx context.Context, source *Config) error
	// DeleteSource deletes a source by name.
	DeleteSource(ctx context.Context, name string) error
	// ValidateSource validates a source configuration and ensures required indices exist.
	ValidateSource(
		ctx context.Context,
		sourceName string,
		indexManager storagetypes.IndexManager,
	) (*configtypes.Source, error)
	// GetMetrics returns the current metrics.
	GetMetrics() types.SourcesMetrics
	// FindByName finds a source by name. Returns nil if not found.
	FindByName(name string) *Config
	// GetSources retrieves all source configurations.
	GetSources() ([]Config, error)
}

// Params contains the parameters for creating a new source manager.
type Params struct {
	// Logger is the logger to use.
	Logger logger.Interface
}

// ErrInvalidSource is returned when a source is invalid.
var ErrInvalidSource = errors.New("invalid source")

// ErrSourceNotFound is returned when a source is not found.
var ErrSourceNotFound = errors.New("source not found")

// ErrSourceExists is returned when a source already exists.
var ErrSourceExists = errors.New("source already exists")

// ValidateParams validates the parameters for creating a new source manager.
func ValidateParams(p Params) error {
	if p.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}

// Config represents a source configuration.
type Config = types.SourceConfig

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig = types.SelectorConfig

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors = types.ArticleSelectors

// Sources manages a collection of web content sources.
type Sources struct {
	sources []Config
	logger  logger.Interface
	metrics *types.SourcesMetrics
}

// Ensure Sources implements Interface
var _ Interface = (*Sources)(nil)

// convertArticleSelectors converts article selectors from various types to types.ArticleSelectors.
func convertArticleSelectors(s any) types.ArticleSelectors {
	switch v := s.(type) {
	case configtypes.ArticleSelectors:
		return types.ArticleSelectors{
			Container:     v.Container,
			Title:         v.Title,
			Body:          v.Body,
			Intro:         v.Intro,
			Link:          v.Link,
			Image:         v.Image,
			Byline:        v.Byline,
			PublishedTime: v.PublishedTime,
			TimeAgo:       v.TimeAgo,
			JSONLD:        v.JSONLD,
			Section:       v.Section,
			Keywords:      v.Keywords,
			Description:   v.Description,
			OGTitle:       v.OGTitle,
			OGDescription: v.OGDescription,
			OGImage:       v.OGImage,
			OGType:        v.OGType,
			OGSiteName:    v.OGSiteName,
			OgURL:         v.OgURL,
			Canonical:     v.Canonical,
			WordCount:     v.WordCount,
			PublishDate:   v.PublishDate,
			Category:      v.Category,
			Tags:          v.Tags,
			Author:        v.Author,
			BylineName:    v.BylineName,
			ArticleID:     v.ArticleID,
			Exclude:       v.Exclude,
		}
	case loader.ArticleSelectors:
		return types.ArticleSelectors{
			Container:     v.Container,
			Title:         v.Title,
			Body:          v.Body,
			Intro:         v.Intro,
			Link:          v.Link,
			Image:         v.Image,
			Byline:        v.Byline,
			PublishedTime: v.PublishedTime,
			TimeAgo:       v.TimeAgo,
			JSONLD:        v.JSONLD,
			Section:       v.Section,
			Keywords:      v.Keywords,
			Description:   v.Description,
			OGTitle:       v.OGTitle,
			OGDescription: v.OGDescription,
			OGImage:       v.OGImage,
			OGType:        v.OGType,
			OGSiteName:    v.OGSiteName,
			OgURL:         v.OgURL,
			Canonical:     v.Canonical,
			WordCount:     v.WordCount,
			PublishDate:   v.PublishDate,
			Category:      v.Category,
			Tags:          v.Tags,
			Author:        v.Author,
			BylineName:    v.BylineName,
			ArticleID:     v.ArticleID,
			Exclude:       v.Exclude,
		}
	default:
		return types.ArticleSelectors{}
	}
}

// convertListSelectors converts list selectors from various types to types.ListSelectors.
func convertListSelectors(s any) types.ListSelectors {
	switch v := s.(type) {
	case configtypes.ListSelectors:
		return types.ListSelectors{
			Container:       v.Container,
			ArticleCards:    v.ArticleCards,
			ArticleList:     v.ArticleList,
			ExcludeFromList: v.ExcludeFromList,
		}
	case loader.ListSelectors:
		return types.ListSelectors{
			Container:       v.Container,
			ArticleCards:    v.ArticleCards,
			ArticleList:     v.ArticleList,
			ExcludeFromList: v.ExcludeFromList,
		}
	default:
		return types.ListSelectors{}
	}
}

// convertPageSelectors converts page selectors from various types to types.PageSelectors.
func convertPageSelectors(s any) types.PageSelectors {
	switch v := s.(type) {
	case configtypes.PageSelectors:
		return types.PageSelectors{
			Container:     v.Container,
			Title:         v.Title,
			Content:       v.Content,
			Description:   v.Description,
			Keywords:      v.Keywords,
			OGTitle:       v.OGTitle,
			OGDescription: v.OGDescription,
			OGImage:       v.OGImage,
			OgURL:         v.OgURL,
			Canonical:     v.Canonical,
			Exclude:       v.Exclude,
		}
	case loader.PageSelectors:
		return types.PageSelectors{
			Container:     v.Container,
			Title:         v.Title,
			Content:       v.Content,
			Description:   v.Description,
			Keywords:      v.Keywords,
			OGTitle:       v.OGTitle,
			OGDescription: v.OGDescription,
			OGImage:       v.OGImage,
			OgURL:         v.OgURL,
			Canonical:     v.Canonical,
			Exclude:       v.Exclude,
		}
	default:
		return types.PageSelectors{}
	}
}

// createSelectorConfig creates a new SelectorConfig from the given selectors.
func createSelectorConfig(selectors any) types.SelectorConfig {
	var articleSelectors types.ArticleSelectors
	var listSelectors types.ListSelectors
	var pageSelectors types.PageSelectors

	switch s := selectors.(type) {
	case configtypes.SourceSelectors:
		articleSelectors = convertArticleSelectors(s.Article)
		listSelectors = convertListSelectors(s.List)
		pageSelectors = convertPageSelectors(s.Page)
	case configtypes.ArticleSelectors:
		articleSelectors = convertArticleSelectors(s)
	case loader.SourceSelectors:
		articleSelectors = convertArticleSelectors(s.Article)
		listSelectors = convertListSelectors(s.List)
		pageSelectors = convertPageSelectors(s.Page)
	case loader.ArticleSelectors:
		articleSelectors = convertArticleSelectors(s)
	default:
		// Return empty selectors for unknown types
		articleSelectors = types.ArticleSelectors{}
		listSelectors = types.ListSelectors{}
		pageSelectors = types.PageSelectors{}
	}

	return types.SelectorConfig{
		Article: articleSelectors,
		List:    listSelectors,
		Page:    pageSelectors,
	}
}

// LoadSources creates a new Sources instance by loading sources from either the
// gosources API or a YAML file specified in the crawler config. Returns an error if no sources are found.
// The logger parameter is optional and can be nil.
func LoadSources(cfg config.Interface, log logger.Interface) (*Sources, error) {
	var sources []Config
	var err error

	crawlerCfg := cfg.GetCrawlerConfig()

	// API loader is required - no file-based fallback
	if crawlerCfg == nil || crawlerCfg.SourcesAPIURL == "" {
		return nil, errors.New("sources_api_url is required in crawler configuration. API-only mode is enabled")
	}

	if log != nil {
		log.Info("Loading sources from API", "url", crawlerCfg.SourcesAPIURL)
	}
	sources, err = loadSourcesFromAPI(crawlerCfg.SourcesAPIURL, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load sources from API: %w", err)
	}

	// If no sources found, return an error
	if len(sources) == 0 {
		return nil, errors.New("no sources found")
	}

	return &Sources{
		sources: sources,
		logger:  log,
		metrics: types.NewSourcesMetrics(),
	}, nil
}

// loadSourcesFromAPI attempts to load sources from the gosources API
func loadSourcesFromAPI(apiURL string, log logger.Interface) ([]Config, error) {
	apiLoader := loader.NewAPILoader(apiURL, log)

	configs, err := apiLoader.LoadSources()
	if err != nil {
		return nil, fmt.Errorf("failed to load sources from API: %w", err)
	}

	if len(configs) == 0 {
		return nil, errors.New("no sources found from API")
	}

	// Convert loaded configs to our source type
	sourceConfigs := make([]Config, 0, len(configs))
	for i := range configs {
		sourceConfigs = append(sourceConfigs, convertLoaderConfig(configs[i]))
	}

	return sourceConfigs, nil
}

// convertLoaderConfig converts a loader.Config to a types.SourceConfig
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
			PageIndex:      cfg.PageIndex,
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
		PageIndex:      cfg.PageIndex,
		Selectors:      createSelectorConfig(cfg.Selectors),
		Rules:          configtypes.Rules{},
	}
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

	// Find the requested source (case-insensitive match)
	var selectedSource *Config
	var availableNames []string
	for i := range sourceConfigs {
		availableNames = append(availableNames, sourceConfigs[i].Name)
		// Try exact match first
		if sourceConfigs[i].Name == sourceName {
			selectedSource = &sourceConfigs[i]
			break
		}
	}

	// If exact match not found, try case-insensitive match
	if selectedSource == nil {
		for i := range sourceConfigs {
			if strings.EqualFold(sourceConfigs[i].Name, sourceName) {
				selectedSource = &sourceConfigs[i]
				break
			}
		}
	}

	// If source not found, return an error with available sources
	if selectedSource == nil {
		return nil, fmt.Errorf("source not found: %s. Available sources: %v", sourceName, availableNames)
	}

	// Convert to configtypes.Source
	source := types.ConvertToConfigSource(selectedSource)

	// Ensure article index exists if specified
	if selectedSource.ArticleIndex != "" {
		if indexErr := indexManager.EnsureArticleIndex(ctx, selectedSource.ArticleIndex); indexErr != nil {
			return nil, fmt.Errorf("failed to ensure article index exists: %w", indexErr)
		}
	}

	// Ensure page index exists if specified
	// Use PageIndex if available, fallback to Index for backward compatibility
	pageIndexName := selectedSource.PageIndex
	if pageIndexName == "" {
		pageIndexName = selectedSource.Index
	}
	if pageIndexName != "" {
		if pageErr := indexManager.EnsurePageIndex(ctx, pageIndexName); pageErr != nil {
			return nil, fmt.Errorf("failed to ensure page index exists: %w", pageErr)
		}
	}

	return source, nil
}

// GetMetrics returns the current metrics.
func (s *Sources) GetMetrics() types.SourcesMetrics {
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
