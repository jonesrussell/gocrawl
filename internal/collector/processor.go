package collector

import (
	"github.com/gocolly/colly/v2"
)

// ContentProcessor processes non-article content
type ContentProcessor interface {
	ProcessContent(e *colly.HTMLElement)
}

// Constants for default selectors
const (
	// Default selectors when none are specified in the source config
	DefaultArticleSelector    = "article, .article, .post, .entry, .content" // Common article selectors
	DefaultTitleSelector      = "h1"
	DefaultDateSelector       = "time"
	DefaultAuthorSelector     = "span.author"
	DefaultCategoriesSelector = "div.categories"
)

// getSelector returns the specified selector or falls back to a default
func getSelector(specified, defaultSelector string) string {
	if specified == "" {
		return defaultSelector
	}
	return specified
}
