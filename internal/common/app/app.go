// Package app provides common application functionality for GoCrawl commands.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
)

const (
	// DefaultChannelBufferSize is the default size for buffered channels.
	DefaultChannelBufferSize = 100
)

// SetupCollector creates and configures a new collector for the given source.
func SetupCollector(
	ctx context.Context,
	log common.Logger,
	source sources.Config,
	processors []models.ContentProcessor,
	done chan struct{},
	cfg config.Interface,
) (collector.Result, error) {
	rateLimit, err := time.ParseDuration(source.RateLimit)
	if err != nil {
		return collector.Result{}, fmt.Errorf("invalid rate limit format: %w", err)
	}

	// Convert source config to the expected type
	sourceConfig := common.ConvertSourceConfig(&source)
	if sourceConfig == nil {
		return collector.Result{}, errors.New("source configuration is nil")
	}

	// Extract domain from source URL
	domain, err := common.ExtractDomain(source.URL)
	if err != nil {
		return collector.Result{}, fmt.Errorf("error extracting domain: %w", err)
	}

	return collector.New(collector.Params{
		BaseURL:          source.URL,
		MaxDepth:         source.MaxDepth,
		RateLimit:        rateLimit,
		Logger:           log,
		Context:          ctx,
		ArticleProcessor: processors[0], // First processor handles articles
		ContentProcessor: processors[1], // Second processor handles content
		Done:             done,
		Debugger:         logger.NewCollyDebugger(log),
		Source:           sourceConfig,
		Parallelism:      cfg.GetCrawlerConfig().Parallelism,
		RandomDelay:      cfg.GetCrawlerConfig().RandomDelay,
		AllowedDomains:   []string{domain},
	})
}

// ConfigureCrawler sets up the crawler with the given source configuration.
func ConfigureCrawler(c interface {
	SetCollector(*colly.Collector)
	SetMaxDepth(int)
	SetRateLimit(string) error
}, source sources.Config, collector collector.Result) error {
	c.SetCollector(collector.Collector)
	c.SetMaxDepth(source.MaxDepth)
	if err := c.SetRateLimit(source.RateLimit); err != nil {
		return fmt.Errorf("error setting rate limit: %w", err)
	}
	return nil
}

// GracefulShutdown performs a graceful shutdown of the provided fx.App.
// It creates a timeout context and handles any shutdown errors.
type Shutdowner interface {
	Stop(context.Context) error
}
