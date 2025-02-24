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

// ProcessPage handles article extraction
func (p *Processor) ProcessPage(e *colly.HTMLElement) {
	p.Logger.Debug("Processing page", "url", e.Request.URL.String())
	article := p.ArticleService.ExtractArticle(e)
	if article == nil {
		p.Logger.Debug("No article extracted", "url", e.Request.URL.String())
		return
	}
	p.Logger.Debug("Article extracted", "url", e.Request.URL.String(), "title", article.Title)

	// Use the dynamic index name from the Processor instance
	if err := p.Storage.IndexDocument(context.Background(), p.IndexName, article.ID, article); err != nil {
		p.Logger.Error("Failed to index article", "articleID", article.ID, "error", err)
		return
	}

	p.ArticleChan <- article
}
