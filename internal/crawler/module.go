package crawler

import (
	"errors"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// ProvideCrawler creates a new Crawler instance
func ProvideCrawler(p Params) (*Crawler, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	if p.Config == nil {
		return nil, errors.New("config is required")
	}

	// Log the entire configuration to ensure it's set correctly
	p.Logger.Debug("Initializing Crawler Configuration", "config", p.Config)

	crawler := &Crawler{
		Storage:        p.Storage,
		Logger:         p.Logger,
		Debugger:       p.Debugger,
		IndexName:      p.Config.Crawler.IndexName,
		articleChan:    make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewService(p.Logger),
		IndexSvc:       storage.NewIndexService(p.Logger),
		Config:         p.Config,
	}

	return crawler, nil
}

// Module provides the crawler module and its dependencies
var Module = fx.Module("crawler",
	fx.Provide(
		provideCollyDebugger,
		ProvideCrawler,
	),
)

// Params holds the dependencies required to create a new Crawler instance
type Params struct {
	fx.In

	Logger   logger.Interface
	Storage  storage.Interface
	Debugger *logger.CollyDebugger
	Config   *config.Config
}

// Result holds the result of creating a new Crawler instance
type Result struct {
	fx.Out

	Crawler *Crawler
}

// Crawler represents a web crawler
type Crawler struct {
	Storage        storage.Interface
	Collector      *colly.Collector // This will be set later
	Logger         logger.Interface
	Debugger       *logger.CollyDebugger
	IndexName      string
	articleChan    chan *models.Article
	ArticleService article.Interface
	IndexSvc       storage.IndexServiceInterface
	Config         *config.Config
}

const (
	DefaultMaxDepth    = 3
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultParallelism = 2
	DefaultBatchSize   = 100
)
