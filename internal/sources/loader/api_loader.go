// Package loader provides functionality for loading source configurations.
package loader

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources/apiclient"
)

// APILoader handles loading source configurations from the gosources API.
type APILoader struct {
	client *apiclient.Client
	logger logger.Interface
}

// NewAPILoader creates a new APILoader instance.
func NewAPILoader(apiURL string, log logger.Interface) *APILoader {
	client := apiclient.NewClient(apiclient.WithBaseURL(apiURL))
	return &APILoader{
		client: client,
		logger: log,
	}
}

// LoadSources loads all sources from the gosources API.
func (l *APILoader) LoadSources() ([]Config, error) {
	ctx := context.Background()

	apiSources, err := l.client.ListSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources from API: %w", err)
	}

	if len(apiSources) == 0 {
		return nil, ErrNoSources
	}

	// Convert API sources to Config structs
	configs := make([]Config, 0, len(apiSources))
	for i := range apiSources {
		cfg, convertErr := convertAPISourceToConfig(&apiSources[i])
		if convertErr != nil {
			// Log the error but continue processing other sources
			sourceName := "unknown"
			if apiSources[i].Name != "" {
				sourceName = apiSources[i].Name
			} else if apiSources[i].URL != "" {
				sourceName = apiSources[i].URL
			}
			if l.logger != nil {
				l.logger.Warn("Failed to convert source from API, skipping",
					"source", sourceName,
					"error", convertErr,
				)
			}
			continue
		}
		configs = append(configs, cfg)
	}

	if len(configs) == 0 {
		return nil, ErrNoSources
	}

	return configs, nil
}

// convertAPISourceToConfig converts an apiclient.APISource to a loader.Config.
func convertAPISourceToConfig(apiSource *apiclient.APISource) (Config, error) {
	if apiSource == nil {
		return Config{}, errors.New("apiSource cannot be nil")
	}

	return Config{
		Name:         apiSource.Name,
		URL:          apiSource.URL,
		RateLimit:    apiSource.RateLimit,
		MaxDepth:     apiSource.MaxDepth,
		Time:         apiSource.Time,
		ArticleIndex: apiSource.ArticleIndex,
		PageIndex:    apiSource.PageIndex,
		Index:        apiSource.PageIndex, // For backward compatibility
		Selectors: SourceSelectors{
			Article: convertAPIArticleSelectors(apiSource.Selectors.Article),
			List:    convertAPIListSelectors(apiSource.Selectors.List),
			Page:    convertAPIPageSelectors(apiSource.Selectors.Page),
		},
	}, nil
}

// convertAPIArticleSelectors converts API article selectors to loader article selectors.
func convertAPIArticleSelectors(api apiclient.APIArticleSelectors) ArticleSelectors {
	return ArticleSelectors{
		Container:     api.Container,
		Title:         api.Title,
		Body:          api.Body,
		Intro:         api.Intro,
		Link:          api.Link,
		Image:         api.Image,
		Byline:        api.Byline,
		PublishedTime: api.PublishedTime,
		TimeAgo:       api.TimeAgo,
		JSONLD:        api.JSONLD,
		Section:       api.Section,
		Keywords:      api.Keywords,
		Description:   api.Description,
		OGTitle:       api.OGTitle,
		OGDescription: api.OGDescription,
		OGImage:       api.OGImage,
		OGType:        api.OGType,
		OGSiteName:    api.OGSiteName,
		OgURL:         api.OgURL,
		Canonical:     api.Canonical,
		Category:      api.Category,
		Author:        api.Author,
		ArticleID:     api.ArticleID,
		Exclude:       api.Exclude,
	}
}

// convertAPIListSelectors converts API list selectors to loader list selectors.
func convertAPIListSelectors(api apiclient.APIListSelectors) ListSelectors {
	return ListSelectors{
		Container:       api.Container,
		ArticleCards:    api.ArticleCards,
		ArticleList:     api.ArticleList,
		ExcludeFromList: api.ExcludeFromList,
	}
}

// convertAPIPageSelectors converts API page selectors to loader page selectors.
func convertAPIPageSelectors(api apiclient.APIPageSelectors) PageSelectors {
	return PageSelectors{
		Container:     api.Container,
		Title:         api.Title,
		Content:       api.Content,
		Description:   api.Description,
		Keywords:      api.Keywords,
		OGTitle:       api.OGTitle,
		OGDescription: api.OGDescription,
		OGImage:       api.OGImage,
		OgURL:         api.OgURL,
		Canonical:     api.Canonical,
		Exclude:       api.Exclude,
	}
}
