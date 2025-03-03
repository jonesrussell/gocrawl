package article

import (
	"context"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
)

// Processor is responsible for processing articles
type Processor struct {
	Logger         logger.Interface
	ArticleService Interface
	Storage        storage.Interface
	IndexName      string
	ArticleChan    chan *models.Article
}

// ProcessArticle handles article extraction
func (p *Processor) ProcessArticle(e *colly.HTMLElement) {
	p.Logger.Debug("Processing article",
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

	// Use the dynamic index name from the Processor instance
	if err := p.Storage.IndexDocument(context.Background(), p.IndexName, article.ID, article); err != nil {
		p.Logger.Error("Failed to index article",
			"component", "article/processor",
			"articleID", article.ID,
			"error", err)
		return
	}

	p.ArticleChan <- article
}

// ProcessContent implements the models.ContentProcessor interface
func (p *Processor) ProcessContent(e *colly.HTMLElement) {
	// For article processor, we'll just call ProcessArticle
	p.ProcessArticle(e)
}

// Ensure Processor implements models.ContentProcessor
var _ models.ContentProcessor = (*Processor)(nil)
