package crawler

import (
	"context"
	"errors"

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
}

func provideCollyDebugger(log logger.Interface) *logger.CollyDebugger {
	return logger.NewCollyDebugger(log)
}

// ProvideCrawler creates a new Crawler instance
func ProvideCrawler(p Params) (Interface, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required")
	}

	if p.Config == nil {
		return nil, errors.New("config is required")
	}

	// Log the entire configuration to ensure it's set correctly
	p.Logger.Debug("Initializing Crawler Configuration", "config", p.Config)

	// Create a new crawler instance
	crawler := &Crawler{
		Storage:        p.Storage,
		Logger:         p.Logger,
		Debugger:       p.Debugger,
		IndexName:      p.Config.Crawler.IndexName,
		articleChan:    make(chan *models.Article, DefaultBatchSize),
		ArticleService: article.NewService(p.Logger),
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

	Crawler Interface
}

const (
	DefaultMaxDepth    = 3
	DefaultMaxBodySize = 10 * 1024 * 1024 // 10 MB
	DefaultParallelism = 2
	DefaultBatchSize   = 100
)
