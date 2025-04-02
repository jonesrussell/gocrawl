// Package app provides common application functionality for GoCrawl commands.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	mockutils "github.com/jonesrussell/gocrawl/internal/testutils"
	"go.uber.org/fx"
)

const (
	// DefaultChannelBufferSize is the default size for buffered channels.
	DefaultChannelBufferSize = 100
)

// Params holds the parameters for creating an application.
type Params struct {
	fx.In

	Config config.Interface
	Logger common.Logger
}

// SetupCollector creates and configures a new collector for the given source.
func SetupCollector(
	ctx context.Context,
	log common.Logger,
	source sources.Config,
	processors []common.Processor,
	done chan struct{},
	cfg config.Interface,
) (crawler.Interface, error) {
	// Convert source config to the expected type
	sourceConfig := common.ConvertSourceConfig(&source)
	if sourceConfig == nil {
		return nil, errors.New("source configuration is nil")
	}

	// Extract domain from source URL
	domain, err := common.ExtractDomain(source.URL)
	if err != nil {
		return nil, fmt.Errorf("error extracting domain: %w", err)
	}

	// Create a new collector with debug logging
	debugger := &debug.LogDebugger{}
	c := colly.NewCollector(
		colly.MaxDepth(source.MaxDepth),
		colly.Async(true),
		colly.AllowURLRevisit(),
		colly.AllowedDomains(domain),
		colly.Debugger(debugger),
	)

	// Set up rate limiting
	if rateErr := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: source.RateLimit,
		Parallelism: crawler.DefaultParallelism,
	}); rateErr != nil {
		return nil, fmt.Errorf("error setting rate limit: %w", rateErr)
	}

	// Create a new crawler using fx
	var crawlerResult crawler.Result
	app := fx.New(
		fx.Provide(
			fx.Annotate(
				func() common.Logger { return log },
				fx.ResultTags(`name:"logger"`),
			),
			fx.Annotate(
				func() debug.Debugger { return debugger },
				fx.ResultTags(`name:"debugger"`),
			),
			fx.Annotate(
				func() common.Processor { return processors[0] },
				fx.ResultTags(`name:"articleProcessor"`),
			),
			fx.Annotate(
				func() common.Processor { return processors[1] },
				fx.ResultTags(`name:"contentProcessor"`),
			),
			fx.Annotate(
				func() sources.Interface { return &sources.Sources{} },
				fx.ResultTags(`name:"sources"`),
			),
			fx.Annotate(
				events.NewBus,
				fx.ResultTags(`name:"bus"`),
			),
			fx.Annotate(
				func() api.IndexManager { return &mockutils.MockIndexManager{} },
				fx.ResultTags(`name:"indexManager"`),
			),
			crawler.ProvideCrawler,
		),
		fx.Populate(&crawlerResult),
	)

	// Start the application
	if startErr := app.Start(ctx); startErr != nil {
		return nil, fmt.Errorf("error starting crawler: %w", startErr)
	}

	// Configure the crawler
	crawlerResult.Crawler.SetCollector(c)

	return crawlerResult.Crawler, nil
}

// ConfigureCrawler configures a crawler with the given source and collector.
func ConfigureCrawler(c interface {
	SetCollector(*colly.Collector)
	SetMaxDepth(int)
	SetRateLimit(time.Duration) error
}, source sources.Config, crawler crawler.Interface) error {
	if err := c.SetRateLimit(source.RateLimit); err != nil {
		return fmt.Errorf("error setting rate limit: %w", err)
	}
	c.SetMaxDepth(source.MaxDepth)
	return nil
}

// Shutdowner defines the interface for components that need cleanup.
type Shutdowner interface {
	Stop(context.Context) error
}

// Module provides the application module and its dependencies.
var Module = fx.Module("app",
	fx.Provide(
		// Provide configuration
		config.LoadConfig,
		// Provide logger
		func() (common.Logger, error) {
			return logger.NewCustomLogger(nil, logger.Params{
				Debug:  true,
				Level:  "info",
				AppEnv: "development",
			})
		},
		// Provide event bus
		events.NewBus,
		// Provide crawler
		crawler.ProvideCrawler,
	),
)
