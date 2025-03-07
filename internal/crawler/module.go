package crawler

import (
	"context"
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Interface defines the methods required for a crawler
type Interface interface {
	Start(ctx context.Context, url string) error
	Stop()
	SetCollector(collector *colly.Collector)
	SetService(service article.Interface)
	GetBaseURL() string
	GetIndexManager() storage.IndexServiceInterface
}

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// Params holds the dependencies for creating a crawler
type Params struct {
	fx.In

	Logger           logger.Interface
	Storage          storage.Interface
	Debugger         *logger.CollyDebugger
	Config           *config.Config
	Source           string `name:"sourceName"`
	IndexService     storage.IndexServiceInterface
	ContentProcessor []models.ContentProcessor `group:"processors"`
}

// Result holds the crawler instance
type Result struct {
	fx.Out

	Crawler Interface
}

// Module provides crawler-related dependencies
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		ProvideCrawler,
	),
)

// ProvideCrawler creates a new Crawler instance
func ProvideCrawler(p Params) (Interface, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	if p.Config == nil {
		return nil, errors.New("config is required")
	}

	if p.Storage == nil {
		return nil, errors.New("storage is required")
	}

	if p.IndexService == nil {
		return nil, errors.New("index service is required")
	}

	if len(p.ContentProcessor) == 0 {
		return nil, errors.New("at least one content processor is required")
	}

	// Log the entire configuration to ensure it's set correctly
	p.Logger.Debug("Initializing Crawler Configuration", "config", p.Config)

	// Create a new crawler instance
	crawler := &Crawler{
		Storage:     p.Storage,
		Logger:      p.Logger,
		Debugger:    p.Debugger,
		IndexName:   p.Config.Crawler.IndexName,
		articleChan: make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewServiceWithConfig(article.ServiceParams{
			Logger: p.Logger,
			Config: p.Config,
			Source: p.Source,
		}),
		IndexService:     p.IndexService,
		Config:           p.Config,
		ContentProcessor: p.ContentProcessor[0], // Use the first processor
	}

	return crawler, nil
}

const (
	DefaultMaxDepth    = 3
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultParallelism = 2
	DefaultBatchSize   = 100
)
