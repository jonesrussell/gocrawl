package app

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/crawler"
)

// StartCrawler starts the crawler using the main application instance
func StartCrawler(ctx context.Context, crawler *crawler.Crawler) error {
	if err := crawler.Start(ctx); err != nil {
		return err
	}
	return nil
}
