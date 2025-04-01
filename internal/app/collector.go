// Package app provides application-level functionality and setup.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/pkg/collector"
)

// SetupCollector creates and configures a new collector instance.
//
// Parameters:
//   - ctx: The context for the collector
//   - log: The logger to use
//   - source: The source configuration
//   - processors: The content processors to use
//   - done: A channel that signals when crawling is complete
//   - cfg: The application configuration
//
// Returns:
//   - collector.Result: The configured collector
//   - error: Any error that occurred during setup
func SetupCollector(
	ctx context.Context,
	log types.Logger,
	source sources.Config,
	processors []collector.Processor,
	done chan struct{},
	cfg config.Interface,
) (collector.Result, error) {
	// Create collector parameters
	params := collector.Params{
		ArticleProcessor: processors[0], // First processor is for articles
		ContentProcessor: processors[1], // Second processor is for content
		BaseURL:          source.URL,
		Context:          ctx,
		Logger:           log,
		MaxDepth:         source.MaxDepth,
		RateLimit:        source.RateLimit,
		Done:             done,
		AllowedDomains:   []string{source.URL},
	}

	// Create collector
	return collector.New(params)
}

// ConfigureCrawler configures a crawler instance with the given source.
//
// Parameters:
//   - c: The crawler instance to configure
//   - source: The source configuration
//   - result: The collector result
//
// Returns:
//   - error: Any error that occurred during configuration
func ConfigureCrawler(
	c crawler.Interface,
	source sources.Config,
	result collector.Result,
) error {
	if source.URL == "" {
		return errors.New("source URL is required")
	}

	// Set collector
	c.SetCollector(result.Collector)

	// Set rate limit
	if err := c.SetRateLimit(source.RateLimit); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	return nil
}
