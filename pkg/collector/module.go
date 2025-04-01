// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
//
// Key features:
// - Configurable crawling depth and parallelism
// - Rate limiting and politeness delays
// - Content extraction and processing
// - Metrics collection and monitoring
// - Error handling and retries
//
// The package integrates with the article and content packages to process
// different types of web content encountered during crawling.
package collector

import (
	"errors"
	"sync"

	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/pkg/processor"
	"go.uber.org/fx"
)

// Collector defines the interface for the collector module.
type Collector interface {
	// Start begins the crawling process.
	Start(url string) error

	// Stop ends the crawling process.
	Stop() error

	// IsRunning returns whether the collector is currently running.
	IsRunning() bool
}

// ModuleParams holds parameters for creating a collector module.
type ModuleParams struct {
	// Logger is the logger for the collector.
	Logger common.Logger

	// Config is the collector configuration.
	Config *config.Config

	// ArticleProcessor handles article content processing.
	ArticleProcessor processor.Processor

	// ContentProcessor handles general content processing.
	ContentProcessor processor.Processor
}

// collector holds the collector dependencies.
type collector struct {
	// logger is the logger for the collector.
	logger common.Logger

	// config is the collector configuration.
	config *config.Config

	// articleProcessor handles article content processing.
	articleProcessor processor.Processor

	// contentProcessor handles general content processing.
	contentProcessor processor.Processor

	// running indicates whether the collector is currently running.
	running bool

	// mu protects the running field.
	mu sync.RWMutex
}

// NewCollector creates a new collector module.
// It initializes the collector with the provided parameters.
//
// Parameters:
//   - p: ModuleParams containing all required dependencies
//
// Returns:
//   - *collector: The initialized collector module
func NewCollector(p ModuleParams) *collector {
	return &collector{
		logger:           p.Logger,
		config:           p.Config,
		articleProcessor: p.ArticleProcessor,
		contentProcessor: p.ContentProcessor,
	}
}

// Start begins the crawling process.
// It initializes the collector and starts crawling from the given URL.
//
// Parameters:
//   - url: The URL to start crawling from
//
// Returns:
//   - error: Any error that occurred during startup
func (c *collector) Start(url string) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return errors.New("collector is already running")
	}
	c.running = true
	c.mu.Unlock()

	// TODO: Implement actual crawling logic
	return nil
}

// Stop ends the crawling process.
// It ensures all resources are properly cleaned up.
//
// Returns:
//   - error: Any error that occurred during shutdown
func (c *collector) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return errors.New("collector is not running")
	}
	c.running = false
	c.mu.Unlock()

	// TODO: Implement cleanup logic
	return nil
}

// IsRunning returns whether the collector is currently running.
//
// Returns:
//   - bool: Whether the collector is running
func (c *collector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Module provides the collector module for dependency injection.
// It provides:
// - Collector configuration with environment variable support
// - Collector instance with configured settings
// - Metrics collection and monitoring
//
// The module uses fx.Provide to wire up dependencies and ensure proper
// initialization of the collector components. Configuration is loaded from
// environment variables with sensible defaults.
var Module = fx.Module("collector",
	fx.Provide(
		fx.Annotate(
			NewCollector,
			fx.As(new(Collector)),
		),
	),
)
