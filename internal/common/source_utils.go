// Package common provides shared utilities and types used across the application.
package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

// ConvertSourceConfig converts a sources.Config to a config.Source.
// It handles the conversion of fields between the two types.
func ConvertSourceConfig(source *sources.Config) *config.Source {
	if source == nil {
		return nil
	}

	return &config.Source{
		Name:         source.Name,
		URL:          source.URL,
		ArticleIndex: source.ArticleIndex,
		Index:        source.Index,
		RateLimit:    source.RateLimit,
		MaxDepth:     source.MaxDepth,
		Time:         source.Time,
		Selectors: config.SourceSelectors{
			Article: config.ArticleSelectors{
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
