// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// CommandDeps represents the dependencies required by the crawl command.
type CommandDeps struct {
	fx.In

	Context     context.Context `name:"crawlContext"`
	SourceName  string          `name:"sourceName"`
	Logger      logger.Interface
	Crawler     crawler.Interface
	Sources     *sources.Sources
	Handler     *signal.SignalHandler `name:"signalHandler"`
	ArticleChan chan *models.Article  `name:"crawlerArticleChannel"`
	Processors  []common.Processor    `group:"processors"`
}

// CrawlerCommand represents the crawl command.
type CrawlerCommand struct {
	crawler     crawler.Interface
	logger      logger.Interface
	processors  []common.Processor
	articleChan chan *models.Article
	done        chan struct{}
}

// NewCrawlerCommand creates a new crawl command.
func NewCrawlerCommand(
	logger logger.Interface,
	crawler crawler.Interface,
	processors []common.Processor,
	articleChan chan *models.Article,
) *CrawlerCommand {
	return &CrawlerCommand{
		crawler:     crawler,
		logger:      logger,
		processors:  processors,
		articleChan: articleChan,
		done:        make(chan struct{}),
	}
}

// Start starts the crawler.
func (c *CrawlerCommand) Start(ctx context.Context, sourceName string) error {
	// Start crawler
	c.logger.Info("Starting crawler...", "source", sourceName)
	if err := c.crawler.Start(ctx, sourceName); err != nil {
		return fmt.Errorf("failed to start crawler: %w", err)
	}

	// Process articles from the crawler's channel
	go func() {
		defer close(c.done)
		crawlerChan := c.crawler.GetArticleChannel()
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Context cancelled, stopping article processing")
				return
			case article, ok := <-crawlerChan:
				if !ok {
					c.logger.Info("Crawler article channel closed")
					return
				}

				// Forward article to the command's channel
				select {
				case c.articleChan <- article:
					// Article forwarded successfully
				case <-ctx.Done():
					c.logger.Info("Context cancelled while forwarding article")
					return
				}

				// Create a job for each article with timeout
				job := &common.Job{
					ID:        article.ID,
					URL:       article.Source,
					Status:    "pending",
					CreatedAt: article.CreatedAt,
					UpdatedAt: article.UpdatedAt,
				}

				// Process the article using each processor with timeout
				for _, processor := range c.processors {
					// Create a timeout context for each processor
					processorCtx, processorCancel := context.WithTimeout(ctx, 30*time.Second)
					processor.ProcessJob(processorCtx, job)
					processorCancel()
				}
			}
		}
	}()

	return nil
}

// Stop stops the crawler.
func (c *CrawlerCommand) Stop(ctx context.Context) error {
	return c.crawler.Stop(ctx)
}

// Wait waits for the crawler to finish processing.
func (c *CrawlerCommand) Wait() {
	c.crawler.Wait()
	<-c.done
}

// Module provides the crawl command module for dependency injection.
var Module = fx.Module("crawl",
	// Include all required modules
	fx.Provide(
		NewCrawlerCommand,
	),
)
