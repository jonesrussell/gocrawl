// Package sourceutils provides utilities for working with source configurations.
package sourceutils

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/types"
)

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
	Selectors      SelectorConfig
	Rules          types.Rules
}

// SelectorConfig defines the CSS selectors used for content extraction.
type SelectorConfig struct {
	Article ArticleSelectors
}

// ArticleSelectors defines the CSS selectors used for article content extraction.
type ArticleSelectors struct {
	Container     string
	Title         string
	Body          string
	Intro         string
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
	OgURL         string
	Canonical     string
	WordCount     string
	PublishDate   string
	Category      string
	Tags          string
	Author        string
	BylineName    string
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
		RateLimit:      source.RateLimit,
		MaxDepth:       source.MaxDepth,
		Time:           source.Time,
		Index:          source.Index,
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Container:     source.Selectors.Article.Container,
				Title:         source.Selectors.Article.Title,
				Body:          source.Selectors.Article.Body,
				Intro:         source.Selectors.Article.Intro,
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
				OgURL:         source.Selectors.Article.OgURL,
				Canonical:     source.Selectors.Article.Canonical,
				WordCount:     source.Selectors.Article.WordCount,
				PublishDate:   source.Selectors.Article.PublishDate,
				Category:      source.Selectors.Article.Category,
				Tags:          source.Selectors.Article.Tags,
				Author:        source.Selectors.Article.Author,
				BylineName:    source.Selectors.Article.BylineName,
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
