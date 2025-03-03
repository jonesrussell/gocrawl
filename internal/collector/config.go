package collector

import (
	"context"
	"errors"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Context keys
const (
	articleFoundKey = "articleFound"
	bodyElementKey  = "bodyElement"
)

// Params holds the parameters for creating a Collector
type Params struct {
	fx.In

	ArticleProcessor *article.Processor
	ContentProcessor ContentProcessor
	BaseURL          string
	Context          context.Context
	Debugger         *logger.CollyDebugger
	Logger           logger.Interface
	MaxDepth         int
	Parallelism      int
	RandomDelay      time.Duration
	RateLimit        time.Duration
	Source           *sources.Config
}

// Result holds the collector instance
type Result struct {
	fx.Out

	Collector *colly.Collector
}

// ValidateParams validates the collector parameters
func ValidateParams(p Params) error {
	if p.BaseURL == "" {
		return errors.New("base URL cannot be empty")
	}

	if p.ArticleProcessor == nil {
		return errors.New("article processor is required")
	}

	if p.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}
