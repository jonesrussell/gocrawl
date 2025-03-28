// Package article provides functionality for processing and managing articles.
package article

import (
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// ArticleProcessor handles article content processing.
type ArticleProcessor struct {
	// Logger for article processing operations
	Logger logger.Interface
	// ArticleService for article operations
	ArticleService Interface
	// Storage for article persistence
	Storage types.Interface
	// IndexName is the name of the article index
	IndexName string
	// ArticleChan is the channel for sending processed articles
	ArticleChan chan *models.Article
}

// Process implements the collector.Processor interface
func (p *ArticleProcessor) Process(e *colly.HTMLElement) error {
	// Extract article data from HTML element
	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article found in HTML element",
			"component", "article/processor",
			"url", e.Request.URL.String())
		return nil
	}

	// Process the article
	if err := p.ArticleService.Process(article); err != nil {
		p.Logger.Error("Failed to process article",
			"component", "article/processor",
			"articleID", article.ID,
			"error", err)
		return err
	}

	// Send to channel if available
	if p.ArticleChan != nil {
		p.ArticleChan <- article
	}

	return nil
}

// ProcessHTMLElement handles article extraction from HTML elements
func (p *ArticleProcessor) ProcessHTMLElement(e *colly.HTMLElement) {
	p.Logger.Debug("Processing article from HTML",
		"component", "article/processor",
		"url", e.Request.URL.String())

	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article extracted",
			"component", "article/processor",
			"url", e.Request.URL.String())
		return
	}

	p.Logger.Debug("Article extracted",
		"component", "article/processor",
		"url", e.Request.URL.String(),
		"title", article.Title)

	if err := p.ArticleService.Process(article); err != nil {
		p.Logger.Error("Failed to process article",
			"component", "article/processor",
			"articleID", article.ID,
			"error", err)
	}
}

// ProcessContent implements the collector.Processor interface
func (p *ArticleProcessor) ProcessContent(e *colly.HTMLElement) {
	// Skip content pages - we only process articles
	p.Logger.Debug("Skipping content page in article processor",
		"component", "article/processor",
		"url", e.Request.URL.String())
}

// Ensure ArticleProcessor implements collector.Processor
var _ collector.Processor = (*ArticleProcessor)(nil)
