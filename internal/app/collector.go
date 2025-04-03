// Package app provides application-level functionality and setup.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/testutils"
)

// SetupCollector initializes and configures a new collector instance.
//
// Parameters:
//   - ctx: The context for the collector operation
//   - log: The logger instance
//   - source: The source configuration
//   - processors: List of processors to apply to collected data
//   - done: A channel that signals when crawling is complete
//   - cfg: The application configuration
//   - client: The Elasticsearch client
//
// Returns:
//   - crawler.Interface: The configured crawler instance
//   - error: Any error that occurred during setup
func SetupCollector(
	ctx context.Context,
	log logger.Interface,
	source sources.Config,
	processors []common.Processor,
	done chan struct{},
	cfg config.Interface,
	client *elasticsearch.Client,
) (crawler.Interface, error) {
	// Create dependencies
	indexManager := testutils.NewMockIndexManager()
	bus := events.NewBus()
	sources := sources.NewSources(&source, log)

	// Create crawler using the module's provider
	result := crawler.ProvideCrawler(
		log,
		indexManager,
		sources,
		processors[0], // First processor is for articles
		processors[1], // Second processor is for content
		bus,
	)

	// Configure crawler with source settings
	if err := ConfigureCrawler(result.Crawler, source); err != nil {
		return nil, fmt.Errorf("failed to configure crawler: %w", err)
	}

	return result.Crawler, nil
}

// ConfigureCrawler configures a crawler instance with the given source.
//
// Parameters:
//   - c: The crawler instance to configure
//   - source: The source configuration
//
// Returns:
//   - error: Any error that occurred during configuration
func ConfigureCrawler(
	c crawler.Interface,
	source sources.Config,
) error {
	if source.URL == "" {
		return errors.New("source URL is required")
	}

	// Set rate limit
	if err := c.SetRateLimit(source.RateLimit); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Set max depth
	c.SetMaxDepth(source.MaxDepth)

	return nil
}
