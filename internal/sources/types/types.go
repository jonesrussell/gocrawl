// Package types provides type definitions for the sources package.
package types

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/types"
)

// Source defines the interface for data sources.
type Source any

// SourcesMetrics contains metrics about the source manager.
type SourcesMetrics struct {
	// SourceCount is the number of sources.
	SourceCount int64
	// LastUpdated is the time the metrics were last updated.
	LastUpdated time.Time
}

// NewSourcesMetrics creates a new SourcesMetrics instance.
func NewSourcesMetrics() *SourcesMetrics {
	return &SourcesMetrics{
		SourceCount: 0,
		LastUpdated: time.Now(),
	}
}

// SourceConfig represents a source configuration.
type SourceConfig struct {
	Name           string
	URL            string
	AllowedDomains []string
	StartURLs      []string
	RateLimit      time.Duration
	MaxDepth       int
	Time           []string
	Index          string
	ArticleIndex   string
	PageIndex      string
	Selectors      SelectorConfig
	Rules          types.Rules
}

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig struct {
	Article ArticleSelectors
	List    ListSelectors
	Page    PageSelectors
}

// ListSelectors defines the CSS selectors used for article list page extraction.
type ListSelectors struct {
	Container       string
	ArticleCards    string
	ArticleList     string
	ExcludeFromList []string
}

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors struct {
	Container     string
	Title         string
	Body          string
	Intro         string
	Link          string
	Image         string
	Byline        string
	PublishedTime string
	TimeAgo       string
	JSONLD        string
	Section       string
	Keywords      string
	Description   string
	OGTitle       string
	OGDescription string
	OGImage       string
	OGType        string
	OGSiteName    string
	OgURL         string
	Canonical     string
	WordCount     string
	PublishDate   string
	Category      string
	Tags          string
	Author        string
	BylineName    string
	ArticleID     string
	Exclude       []string
}

// PageSelectors defines the CSS selectors used for page content extraction.
type PageSelectors struct {
	Container     string
	Title         string
	Content       string
	Description   string
	Keywords      string
	OGTitle       string
	OGDescription string
	OGImage       string
	OgURL         string
	Canonical     string
	Exclude       []string
}

// ConvertToConfigSource converts a SourceConfig to a types.Source.
func ConvertToConfigSource(source *SourceConfig) *types.Source {
	if source == nil {
		return nil
	}

	return &types.Source{
		Name:           source.Name,
		URL:            source.URL,
		AllowedDomains: source.AllowedDomains,
		StartURLs:      source.StartURLs,
		RateLimit:      source.RateLimit.String(),
		MaxDepth:       source.MaxDepth,
		Time:           source.Time,
		Index:          source.Index,
		ArticleIndex:   source.ArticleIndex,
		PageIndex:      source.PageIndex,
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Container:     source.Selectors.Article.Container,
				Title:         source.Selectors.Article.Title,
				Body:          source.Selectors.Article.Body,
				Intro:         source.Selectors.Article.Intro,
				Link:          source.Selectors.Article.Link,
				Image:         source.Selectors.Article.Image,
				Byline:        source.Selectors.Article.Byline,
				PublishedTime: source.Selectors.Article.PublishedTime,
				TimeAgo:       source.Selectors.Article.TimeAgo,
				JSONLD:        source.Selectors.Article.JSONLD,
				Section:       source.Selectors.Article.Section,
				Keywords:      source.Selectors.Article.Keywords,
				Description:   source.Selectors.Article.Description,
				OGTitle:       source.Selectors.Article.OGTitle,
				OGDescription: source.Selectors.Article.OGDescription,
				OGImage:       source.Selectors.Article.OGImage,
				OGType:        source.Selectors.Article.OGType,
				OGSiteName:    source.Selectors.Article.OGSiteName,
				OgURL:         source.Selectors.Article.OgURL,
				Canonical:     source.Selectors.Article.Canonical,
				WordCount:     source.Selectors.Article.WordCount,
				PublishDate:   source.Selectors.Article.PublishDate,
				Category:      source.Selectors.Article.Category,
				Tags:          source.Selectors.Article.Tags,
				Author:        source.Selectors.Article.Author,
				BylineName:    source.Selectors.Article.BylineName,
				ArticleID:     source.Selectors.Article.ArticleID,
				Exclude:       source.Selectors.Article.Exclude,
			},
			List: types.ListSelectors{
				Container:       source.Selectors.List.Container,
				ArticleCards:    source.Selectors.List.ArticleCards,
				ArticleList:     source.Selectors.List.ArticleList,
				ExcludeFromList: source.Selectors.List.ExcludeFromList,
			},
			Page: types.PageSelectors{
				Container:     source.Selectors.Page.Container,
				Title:         source.Selectors.Page.Title,
				Content:       source.Selectors.Page.Content,
				Description:   source.Selectors.Page.Description,
				Keywords:      source.Selectors.Page.Keywords,
				OGTitle:       source.Selectors.Page.OGTitle,
				OGDescription: source.Selectors.Page.OGDescription,
				OGImage:       source.Selectors.Page.OGImage,
				OgURL:         source.Selectors.Page.OgURL,
				Canonical:     source.Selectors.Page.Canonical,
				Exclude:       source.Selectors.Page.Exclude,
			},
		},
		Rules: source.Rules,
	}
}

// ExtractDomain extracts the domain from a URL string.
// It handles both full URLs and path-only URLs.
func ExtractDomain(sourceURL string) (string, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Host == "" {
		// If no host in URL, treat the first path segment as the domain
		parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(parts) > 0 {
			return parts[0], nil
		}
		return "", fmt.Errorf("could not extract domain from path: %s", sourceURL)
	}

	return parsedURL.Host, nil
}
