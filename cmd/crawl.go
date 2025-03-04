package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var sourceName string

// createCrawlCmd creates the crawl command
func createCrawlCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "crawl",
		Short: "Crawl a single source defined in sources.yml",

		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			globalLogger.Debug("Starting crawl-crawl command...", "sourceName", sourceName)

			// Create an Fx application
			app := fx.New(
				fx.Provide(
					func() *config.Config {
						return globalConfig
					},
					func() logger.Interface {
						return globalLogger
					},
					func() chan *models.Article {
						return make(chan *models.Article, 100)
					},
					func() chan *models.Content {
						return make(chan *models.Content, 100)
					},
					// Provide source name
					fx.Annotate(
						func() string {
							return sourceName
						},
						fx.ResultTags(`name:"sourceName"`),
					),
					// Provide ArticleIndex name
					fx.Annotate(
						func(s *sources.Sources) string {
							source, _ := s.FindByName(sourceName)
							if source != nil {
								return source.ArticleIndex
							}
							return ""
						},
						fx.ResultTags(`name:"indexName"`),
					),
					// Provide ContentIndex name
					fx.Annotate(
						func(s *sources.Sources) string {
							source, _ := s.FindByName(sourceName)
							if source != nil {
								return source.Index
							}
							return ""
						},
						fx.ResultTags(`name:"contentIndex"`),
					),
					// Provide article processor first (will be first in the slice)
					fx.Annotate(
						func(
							logger logger.Interface,
							service article.Interface,
							storage storage.Interface,
							params struct {
								fx.In
								IndexName string `name:"indexName"`
							},
						) models.ContentProcessor {
							return &article.Processor{
								Logger:         logger,
								ArticleService: service,
								Storage:        storage,
								IndexName:      params.IndexName,
							}
						},
						fx.ResultTags(`group:"processors"`),
					),
					// Provide content processor second (will be second in the slice)
					fx.Annotate(
						func(
							service content.Interface,
							storage storage.Interface,
							logger logger.Interface,
							params struct {
								fx.In
								IndexName string `name:"contentIndex"`
							},
						) models.ContentProcessor {
							return content.NewProcessor(service, storage, logger, params.IndexName)
						},
						fx.ResultTags(`group:"processors"`),
					),
				),
				storage.Module,
				crawler.Module,
				article.Module,
				content.Module,
				sources.Module,
				fx.Invoke(startCrawl),
			)

			// Start the application
			if err := app.Start(ctx); err != nil {
				return fmt.Errorf("error starting application: %w", err)
			}

			defer func() {
				if err := app.Stop(ctx); err != nil {
					globalLogger.Error("Error stopping application", "context", ctx, "error", err)
				}
			}()

			return nil
		},
	}
}

// CrawlParams holds the parameters for crawl-source crawl
type CrawlParams struct {
	fx.In

	Sources         *sources.Sources
	CrawlerInstance crawler.Interface
	Processors      []models.ContentProcessor `group:"processors"`
}

// startCrawl starts the crawl-source crawl
func startCrawl(p CrawlParams) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Get the crawler instance to access index service
	crawler, ok := p.CrawlerInstance.(*crawler.Crawler)
	if !ok {
		return fmt.Errorf("crawler instance is not of type *crawler.Crawler")
	}

	// Set both the crawler and index manager in the sources
	p.Sources.SetCrawler(p.CrawlerInstance)
	p.Sources.SetIndexManager(crawler.IndexService)

	source, err := p.Sources.FindByName(sourceName)
	if err != nil {
		return err
	}

	// Parse rate limit
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		return fmt.Errorf("invalid rate limit: %w", err)
	}

	// Create the collector using the collector module
	collectorResult, err := collector.New(collector.Params{
		BaseURL:          source.URL,
		MaxDepth:         source.MaxDepth,
		RateLimit:        rateLimit,
		Debugger:         logger.NewCollyDebugger(globalLogger),
		Logger:           globalLogger,
		ArticleProcessor: p.Processors[0], // Use first processor as article processor
		ContentProcessor: p.Processors[1], // Use second processor as content processor
		Source:           source,
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	crawler.SetCollector(collectorResult.Collector)

	// Start the crawl
	return p.Sources.Start(context.Background(), sourceName)
}

func init() {
	crawlCmd := createCrawlCmd()
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.Flags().StringVar(&sourceName, "source", "", "Specify the source to crawl")
	if err := crawlCmd.MarkFlagRequired("source"); err != nil {
		globalLogger.Error("Error marking source flag as required", "error", err)
	}
}
