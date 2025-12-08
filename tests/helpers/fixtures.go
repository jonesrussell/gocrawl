// Package helpers provides testing utilities for integration tests.
package helpers

import (
	"time"

	configtypes "github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/domain"
	sourcestypes "github.com/jonesrussell/gocrawl/internal/sources/types"
)

const (
	// DefaultTestMaxDepth is the default max depth for test sources.
	DefaultTestMaxDepth = 2
	// MinURLLength is the minimum URL length for domain extraction.
	MinURLLength = 8 // "https://" is 8 characters
)

// SourceOption is a function that modifies a source configuration.
type SourceOption func(*sourcestypes.SourceConfig)

// ArticleOption is a function that modifies an article.
type ArticleOption func(*domain.Article)

// PageOption is a function that modifies a page.
type PageOption func(*domain.Page)

// TestSource creates a test source configuration.
func TestSource(name, url string, opts ...SourceOption) *sourcestypes.SourceConfig {
	source := &sourcestypes.SourceConfig{
		Name:           name,
		URL:            url,
		AllowedDomains: []string{extractDomain(url)},
		StartURLs:      []string{url},
		RateLimit:      1 * time.Second,
		MaxDepth:       DefaultTestMaxDepth,
		Index:          "test_pages",
		ArticleIndex:   "test_articles",
		PageIndex:      "test_pages",
		Selectors: sourcestypes.SelectorConfig{
			Article: sourcestypes.ArticleSelectors{
				Container:     "article",
				Title:         "h1",
				Body:          ".content",
				Intro:         ".intro",
				PublishedTime: "time",
			},
			Page: sourcestypes.PageSelectors{
				Container: "main",
				Title:     "h1",
				Content:   ".content",
			},
		},
		Rules: configtypes.Rules{
			{
				Pattern:  url + "/*",
				Action:   "allow",
				Priority: 1,
			},
		},
	}

	for _, opt := range opts {
		opt(source)
	}

	return source
}

// WithMaxDepth sets the max depth for a test source.
func WithMaxDepth(depth int) SourceOption {
	return func(s *sourcestypes.SourceConfig) {
		s.MaxDepth = depth
	}
}

// WithStartURLs sets the start URLs for a test source.
func WithStartURLs(urls []string) SourceOption {
	return func(s *sourcestypes.SourceConfig) {
		s.StartURLs = urls
	}
}

// WithRateLimit sets the rate limit for a test source.
func WithRateLimit(limit time.Duration) SourceOption {
	return func(s *sourcestypes.SourceConfig) {
		s.RateLimit = limit
	}
}

// WithIndexName sets the index name for a test source.
func WithIndexName(index string) SourceOption {
	return func(s *sourcestypes.SourceConfig) {
		s.Index = index
	}
}

// TestArticle creates a test article.
func TestArticle(title, content string, opts ...ArticleOption) *domain.Article {
	now := time.Now()
	article := &domain.Article{
		ID:            generateID(),
		Title:         title,
		Body:          content,
		Author:        "Test Author",
		PublishedDate: now.Add(-24 * time.Hour),
		Source:        "https://example.com/article",
		Tags:          []string{"test", "article"},
		Intro:         "Test article introduction",
		Description:   "Test article description",
		WordCount:     len(content),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for _, opt := range opts {
		opt(article)
	}

	return article
}

// WithArticleID sets the ID for a test article.
func WithArticleID(id string) ArticleOption {
	return func(a *domain.Article) {
		a.ID = id
	}
}

// WithArticleSource sets the source URL for a test article.
func WithArticleSource(source string) ArticleOption {
	return func(a *domain.Article) {
		a.Source = source
	}
}

// WithArticleAuthor sets the author for a test article.
func WithArticleAuthor(author string) ArticleOption {
	return func(a *domain.Article) {
		a.Author = author
	}
}

// WithPublishedDate sets the published date for a test article.
func WithPublishedDate(date time.Time) ArticleOption {
	return func(a *domain.Article) {
		a.PublishedDate = date
	}
}

// TestPage creates a test page.
func TestPage(url, title, content string, opts ...PageOption) *domain.Page {
	now := time.Now()
	page := &domain.Page{
		ID:          generateID(),
		URL:         url,
		Title:       title,
		Content:     content,
		Description: "Test page description",
		Keywords:    []string{"test", "page"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, opt := range opts {
		opt(page)
	}

	return page
}

// WithPageID sets the ID for a test page.
func WithPageID(id string) PageOption {
	return func(p *domain.Page) {
		p.ID = id
	}
}

// WithPageDescription sets the description for a test page.
func WithPageDescription(desc string) PageOption {
	return func(p *domain.Page) {
		p.Description = desc
	}
}

// TestHTMLPage returns HTML content for testing.
func TestHTMLPage(title, body string) string {
	return `<!DOCTYPE html>
<html>
<head>
	<title>` + title + `</title>
	<meta name="description" content="Test page description">
</head>
<body>
	<main>
		<h1>` + title + `</h1>
		<div class="content">` + body + `</div>
	</main>
</body>
</html>`
}

// TestArticleHTML returns HTML content for an article.
func TestArticleHTML(title, body string) string {
	return `<!DOCTYPE html>
<html>
<head>
	<title>` + title + `</title>
	<meta name="description" content="Test article description">
</head>
<body>
	<article>
		<h1>` + title + `</h1>
		<div class="intro">Article introduction</div>
		<time datetime="2024-01-01">January 1, 2024</time>
		<div class="content">` + body + `</div>
	</article>
</body>
</html>`
}

// extractDomain extracts the domain from a URL.
func extractDomain(url string) string {
	// Simple domain extraction - assumes http:// or https://
	if len(url) > MinURLLength-1 {
		start := 0
		if url[:4] == "http" {
			if url[:5] == "https" {
				start = 8
			} else {
				start = 7
			}
		}
		end := start
		for end < len(url) && url[end] != '/' && url[end] != ':' {
			end++
		}
		return url[start:end]
	}
	return "example.com"
}

// generateID generates a simple test ID.
func generateID() string {
	return time.Now().Format("20060102150405") + "-test"
}
