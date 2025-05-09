// Package crawl implements the crawl command for fetching and processing web content.
package crawl

import (
	"context"
	"errors"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	crawlerpkg "github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
	crawlerpkg.Module,

	fx.Provide(
		fx.Annotated{
			Group: "commands",
			Target: func(
				cfg config.Interface,
				log logger.Interface,
				crawler crawlerpkg.Interface,
				bus *events.EventBus,
			) func(parent *cobra.Command) {
				return func(parent *cobra.Command) {
					cmd := &cobra.Command{
						Use:   "crawl [source]",
						Short: "Start the crawler",
						Long: `This command crawls a website for content and stores it in the configured storage.
Specify the source name as an argument.`,
						Args: cobra.ExactArgs(1),
						RunE: func(cmd *cobra.Command, args []string) error {
							// Start crawler
							if err := crawler.Start(cmd.Context(), args[0]); err != nil {
								return err
							}
							return nil
						},
					}

					// Add lifecycle hooks
					fx.Invoke(func(lc fx.Lifecycle) {
						lc.Append(fx.Hook{
							OnStart: func(ctx context.Context) error {
								return nil
							},
							OnStop: func(ctx context.Context) error {
								// Stop crawler
								return crawler.Stop(ctx)
							},
						})
					})

					parent.AddCommand(cmd)
				}
			},
		},
	),
)
