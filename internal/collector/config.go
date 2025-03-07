package collector

import (
	"context"
	"errors"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

// Error messages
const (
	errEmptyBaseURL       = "base URL cannot be empty"
	errMissingArticleProc = "article processor is required"
	errMissingLogger      = "logger is required"
)

// Params holds the parameters for creating a Collector
type Params struct {
	fx.In

	ArticleProcessor models.ContentProcessor
	ContentProcessor models.ContentProcessor
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

// Result holds the collector instance and completion channel
type Result struct {
	fx.Out

	Collector *colly.Collector
	Done      chan struct{} // Channel to signal when crawling is complete
}

// ValidateParams validates the collector parameters
func ValidateParams(p Params) error {
	// Ensure BaseURL is not empty
	if p.BaseURL == "" {
		return errors.New(errEmptyBaseURL)
	}

	// Ensure ArticleProcessor is provided
	if p.ArticleProcessor == nil {
		return errors.New(errMissingArticleProc)
	}

	// Ensure Logger is provided
	if p.Logger == nil {
		return errors.New(errMissingLogger)
	}

	return nil
}
