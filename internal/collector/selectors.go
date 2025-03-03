package collector

import (
	"github.com/jonesrussell/gocrawl/internal/config"
)

// Constants for default selectors
const (
	// Default selectors when none are specified in the source config
	DefaultArticleSelector    = "article, .article"
	DefaultTitleSelector      = "h1, h2"
	DefaultDateSelector       = "time"
	DefaultAuthorSelector     = ".author"
	DefaultCategoriesSelector = "div.categories"
)

// getSelector gets the appropriate selector with fallback
func getSelector(selector interface{}, defaultSelector string) string {
	switch s := selector.(type) {
	case config.ArticleSelectors:
		if s.Container != "" {
			return s.Container
		}
	case string:
		if s != "" {
			return s
		}
	}
	return defaultSelector
}
