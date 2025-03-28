// Package article provides functionality for processing and managing articles.
package article

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/models"
)

// Interface defines the capabilities of the article service.
type Interface interface {
	// ExtractArticle extracts article data from the HTML element
	ExtractArticle(e *colly.HTMLElement) *models.Article
	// Process handles the processing of an article
	Process(article *models.Article) error
}
