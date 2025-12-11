// Package apiclient provides HTTP client functionality for interacting with the gosources API.
package apiclient

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources/types"
)

// ConvertAPISourceToConfig converts an APISource to a types.SourceConfig.
func ConvertAPISourceToConfig(apiSource *APISource) (*types.SourceConfig, error) {
	if apiSource == nil {
		return nil, errors.New("apiSource cannot be nil")
	}

	// Parse rate limit
	rateLimit := time.Second // Default
	if apiSource.RateLimit != "" {
		var err error
		rateLimit, err = time.ParseDuration(apiSource.RateLimit)
		if err != nil {
			return nil, fmt.Errorf("invalid rate limit: %w", err)
		}
	}

	// Parse URL to get domain
	parsedURL, err := url.Parse(apiSource.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Hostname()
	if domain == "" {
		domain = apiSource.URL
	}

	// Set default max depth if not specified
	maxDepth := apiSource.MaxDepth
	if maxDepth == 0 {
		maxDepth = 2
	}

	return &types.SourceConfig{
		Name:           apiSource.Name,
		URL:            apiSource.URL,
		AllowedDomains: []string{domain},
		StartURLs:      []string{apiSource.URL},
		RateLimit:      rateLimit,
		MaxDepth:       maxDepth,
		Time:           apiSource.Time,
		Index:          apiSource.PageIndex, // For backward compatibility
		ArticleIndex:   apiSource.ArticleIndex,
		PageIndex:      apiSource.PageIndex,
		Selectors: types.SelectorConfig{
			Article: convertAPIArticleSelectors(apiSource.Selectors.Article),
			List:    convertAPIListSelectors(apiSource.Selectors.List),
			Page:    convertAPIPageSelectors(apiSource.Selectors.Page),
		},
	}, nil
}

// convertAPIArticleSelectors converts APIArticleSelectors to types.ArticleSelectors.
func convertAPIArticleSelectors(api APIArticleSelectors) types.ArticleSelectors {
	return types.ArticleSelectors{
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

// convertAPIListSelectors converts APIListSelectors to types.ListSelectors.
func convertAPIListSelectors(api APIListSelectors) types.ListSelectors {
	return types.ListSelectors{
		Container:       api.Container,
		ArticleCards:    api.ArticleCards,
		ArticleList:     api.ArticleList,
		ExcludeFromList: api.ExcludeFromList,
	}
}

// convertAPIPageSelectors converts APIPageSelectors to types.PageSelectors.
func convertAPIPageSelectors(api APIPageSelectors) types.PageSelectors {
	return types.PageSelectors{
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

// ConvertConfigToAPISource converts a types.SourceConfig to an APISource.
func ConvertConfigToAPISource(config *types.SourceConfig) *APISource {
	if config == nil {
		return nil
	}

	return &APISource{
		Name:         config.Name,
		URL:          config.URL,
		ArticleIndex: config.ArticleIndex,
		PageIndex:    config.PageIndex,
		RateLimit:    config.RateLimit.String(),
		MaxDepth:     config.MaxDepth,
		Time:         config.Time,
		Enabled:      true,
		Selectors: APISelectors{
			Article: convertArticleSelectorsToAPI(config.Selectors.Article),
			List:    convertListSelectorsToAPI(config.Selectors.List),
			Page:    convertPageSelectorsToAPI(config.Selectors.Page),
		},
	}
}

// convertArticleSelectorsToAPI converts types.ArticleSelectors to APIArticleSelectors.
func convertArticleSelectorsToAPI(sel types.ArticleSelectors) APIArticleSelectors {
	return APIArticleSelectors{
		Container:     sel.Container,
		Title:         sel.Title,
		Body:          sel.Body,
		Intro:         sel.Intro,
		Link:          sel.Link,
		Image:         sel.Image,
		Byline:        sel.Byline,
		PublishedTime: sel.PublishedTime,
		TimeAgo:       sel.TimeAgo,
		Section:       sel.Section,
		Category:      sel.Category,
		ArticleID:     sel.ArticleID,
		JSONLD:        sel.JSONLD,
		Keywords:      sel.Keywords,
		Description:   sel.Description,
		OGTitle:       sel.OGTitle,
		OGDescription: sel.OGDescription,
		OGImage:       sel.OGImage,
		OgURL:         sel.OgURL,
		OGType:        sel.OGType,
		OGSiteName:    sel.OGSiteName,
		Canonical:     sel.Canonical,
		Author:        sel.Author,
		Exclude:       sel.Exclude,
	}
}

// convertListSelectorsToAPI converts types.ListSelectors to APIListSelectors.
func convertListSelectorsToAPI(sel types.ListSelectors) APIListSelectors {
	return APIListSelectors{
		Container:       sel.Container,
		ArticleCards:    sel.ArticleCards,
		ArticleList:     sel.ArticleList,
		ExcludeFromList: sel.ExcludeFromList,
	}
}

// convertPageSelectorsToAPI converts types.PageSelectors to APIPageSelectors.
func convertPageSelectorsToAPI(sel types.PageSelectors) APIPageSelectors {
	return APIPageSelectors{
		Container:     sel.Container,
		Title:         sel.Title,
		Content:       sel.Content,
		Description:   sel.Description,
		Keywords:      sel.Keywords,
		OGTitle:       sel.OGTitle,
		OGDescription: sel.OGDescription,
		OGImage:       sel.OGImage,
		OgURL:         sel.OgURL,
		Canonical:     sel.Canonical,
		Exclude:       sel.Exclude,
	}
}
