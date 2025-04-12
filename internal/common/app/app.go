// Package app provides common application functionality for GoCrawl commands.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerconfig "github.com/jonesrussell/gocrawl/internal/config/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"go.uber.org/fx"
)

const (
	// DefaultChannelBufferSize is the default size for buffered channels.
	DefaultChannelBufferSize = 100
)

// App represents the main application structure.
type App struct {
	// Logger for application operations
	Logger logger.Interface
	// Context for application lifecycle
	Context context.Context
}

// Params contains the parameters for creating a new application.
type Params struct {
	fx.In

	// Logger for application operations
	Logger logger.Interface
	// Context for application lifecycle
	Context context.Context
}

// New creates a new application.
func New(log logger.Interface, ctx context.Context) *App {
	return &App{
		Logger:  log,
		Context: ctx,
	}
}

// SetupCollector creates and configures a new collector for the given source.
func SetupCollector(
	ctx context.Context,
	log logger.Interface,
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
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       source.RateLimit,
		RandomDelay: source.RateLimit / 2,
		Parallelism: crawlerconfig.DefaultParallelism,
	})
	if err != nil {
		return nil, fmt.Errorf("error setting rate limit: %w", err)
	}

	// Create a new crawler using fx
	var crawlerResult crawler.Result
	app := fx.New(
		fx.Provide(
			fx.Annotate(
				func() logger.Interface { return log },
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
		func() (logger.Interface, error) {
			return logger.Constructor(logger.Params{
				Config: &logger.Config{
					Level:       logger.InfoLevel,
					Development: true,
					Encoding:    "console",
				},
			})
		},
		// Provide event bus
		events.NewBus,
		// Provide crawler
		crawler.ProvideCrawler,
	),
)

// ProvideLogger provides a logger for the application.
func ProvideLogger() (logger.Interface, error) {
	return logger.Constructor(logger.Params{
		Config: &logger.Config{
			Level:       logger.InfoLevel,
			Development: true,
			Encoding:    "console",
		},
	})
}

// NewLogger creates a new logger.
func NewLogger() (logger.Interface, error) {
	return logger.Constructor(logger.Params{
		Config: &logger.Config{
			Level:       logger.InfoLevel,
			Development: true,
			Encoding:    "console",
		},
	})
}
