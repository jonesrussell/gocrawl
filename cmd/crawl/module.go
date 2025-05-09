// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerpkg "github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Common errors
var (
	ErrInvalidJob    = errors.New("invalid job")
	ErrInvalidJobURL = errors.New("invalid job URL")
)

// Processor defines the interface for content processors.
type Processor interface {
	Process(ctx context.Context, content any) error
}

const (
	// ArticleChannelBufferSize is the buffer size for the article channel.
	ArticleChannelBufferSize = 100
	// DefaultInitTimeout is the default timeout for module initialization.
	DefaultInitTimeout = 30 * time.Second
)

// Module provides the crawl command module for dependency injection.
var Module = fx.Module("crawl",
	// Include required modules
	storage.Module,
	crawlerpkg.Module,

	fx.Provide(
		// Provide the crawl command
		func(
			cfg config.Interface,
			log logger.Interface,
			crawler crawlerpkg.Interface,
			bus *events.EventBus,
		) *cobra.Command {
			return &cobra.Command{
				Use:   "crawl",
				Short: "Start the crawler",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Create context with dependencies
					ctx := context.Background()
					ctx = context.WithValue(ctx, common.ConfigKey, cfg)
					ctx = context.WithValue(ctx, common.LoggerKey, log)

					// Start crawler
					if err := crawler.Start(ctx, "default"); err != nil {
						return err
					}

					// Wait for crawler to finish
					<-ctx.Done()

					return nil
				},
			}
		},
	),
)

// NewCommand creates a new crawl command.
func NewCommand(p struct {
	fx.In
	Crawler *Crawler
}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawl [source]",
		Short: "Crawl a website",
		Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Crawler.Start(cmd.Context())
		},
	}
	return cmd
}
