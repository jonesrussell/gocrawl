package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	// DefaultChannelBufferSize is the default size for buffered channels
	DefaultChannelBufferSize = 100
)

var sourceName string

// CrawlParams holds the parameters for crawl-source crawl
type CrawlParams struct {
	fx.In

	Lifecycle       fx.Lifecycle
	Sources         *sources.Sources
	CrawlerInstance crawler.Interface
	Processors      []models.ContentProcessor `group:"processors"`
	Logger          common.Logger
	Done            chan struct{} `name:"crawlDone"`
}

// createCrawlCmd creates the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl [source]",
	Short: "Crawl a single source defined in sources.yml",
	Long: `Crawl a single source defined in sources.yml.
The source argument must match a name defined in your sources.yml configuration file.

Example:
  gocrawl crawl example-blog`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			const errMsg = "requires exactly one source name from sources.yml\n\n" +
				"Usage:\n  %s\n\n" +
				"Run 'gocrawl list' to see available sources"
			return fmt.Errorf(errMsg, cmd.UseLine())
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		sourceName = args[0]

		// Set up signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create a channel to receive the done signal
		doneChan := make(chan struct{})

		// Create an Fx application with common module
		app := fx.New(
			fx.Supply(fx.Annotate(
				doneChan,
				fx.ResultTags(`name:"crawlDone"`),
			)),
			fx.Provide(
				func() chan *models.Article {
					return make(chan *models.Article, DefaultChannelBufferSize)
				},
				func() chan *models.Content {
					return make(chan *models.Content, DefaultChannelBufferSize)
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
						logger common.Logger,
						service article.Interface,
						storage common.Storage,
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
						storage common.Storage,
						logger common.Logger,
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
			common.Module,
			crawler.Module,
			article.Module,
			content.Module,
			fx.Invoke(startCrawl),
		)

		// Start the application
		if err := app.Start(ctx); err != nil {
			return fmt.Errorf("error starting application: %w", err)
		}

		// Wait for either signal or completion
		select {
		case sig := <-sigChan:
			common.PrintInfof("\nReceived signal %v, initiating shutdown...", sig)
		case <-ctx.Done():
			common.PrintInfof("\nContext cancelled, initiating shutdown...")
		case <-doneChan:
			common.PrintInfof("\nCrawl completed successfully, initiating shutdown...")
		}

		// Create a context with timeout for graceful shutdown
		stopCtx, stopCancel := context.WithTimeout(context.Background(), common.DefaultOperationTimeout)
		defer stopCancel()

		// Stop the application
		if err := app.Stop(stopCtx); err != nil && !errors.Is(err, context.Canceled) {
			common.PrintErrorf("Error stopping application: %v", err)
			return err
		}

		return nil
	},
}

// startCrawl starts the crawl-source crawl
func startCrawl(p CrawlParams) error {
	if p.CrawlerInstance == nil {
		return errors.New("crawler is not initialized")
	}

	// Get the crawler instance to access index service
	crawler, ok := p.CrawlerInstance.(*crawler.Crawler)
	if !ok {
		return errors.New("crawler instance is not of type *crawler.Crawler")
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
		Debugger:         logger.NewCollyDebugger(p.Logger),
		Logger:           p.Logger,
		ArticleProcessor: p.Processors[0], // Use first processor as article processor
		ContentProcessor: p.Processors[1], // Use second processor as content processor
		Source:           source,
	})
	if err != nil {
		return fmt.Errorf("error creating collector: %w", err)
	}

	// Set the collector in the crawler instance
	crawler.SetCollector(collectorResult.Collector)

	// Use lifecycle hooks to manage the crawl
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start the crawl in a goroutine
			go func() {
				if err := p.Sources.Start(ctx, sourceName); err != nil {
					if !errors.Is(err, context.Canceled) {
						p.Logger.Error("Crawl failed", "error", err)
					}
				}
				// Signal completion
				close(p.Done)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Stopping crawler...")
			p.Sources.Stop()
			return nil
		},
	})

	return nil
}

func init() {
	rootCmd.AddCommand(crawlCmd)
}
