package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests")
	flag.Parse()

	app := createApp(*urlPtr, *maxDepthPtr, *rateLimitPtr)

	// Create a channel to listen for OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Run the application
	if err := app.Start(context.Background()); err != nil {
		log.Fatalf("Application start error: %v", err)
	}

	// Wait for a signal to shutdown
	<-sigs
	log.Println("Received shutdown signal, stopping application...")

	// Stop the application gracefully
	if err := app.Stop(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

func createApp(url string, maxDepth int, rateLimit time.Duration) *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		collector.Module,
		crawler.Module,
		fx.Provide(
			fx.Annotated{
				Name:   "baseURL",
				Target: func() string { return url },
			},
			fx.Annotated{
				Name:   "maxDepth",
				Target: func() int { return maxDepth },
			},
			fx.Annotated{
				Name:   "rateLimit",
				Target: func() time.Duration { return rateLimit },
			},
			func(l *logger.CustomLogger) *logger.CustomDebugger {
				return logger.NewCustomDebugger(l)
			},
		),
		fx.WithLogger(func(l *logger.CustomLogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: l.GetZapLogger()}
		}),
		fx.Invoke(registerHooks),
	)
}

func registerHooks(lc fx.Lifecycle, c *crawler.Crawler, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Println("Starting crawling process...")
			go startCrawling(ctx, c, shutdowner)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutdown process initiated...")
			// No need to call c.Stop() since it has been removed
			return nil
		},
	})
}

func startCrawling(ctx context.Context, c *crawler.Crawler, shutdowner fx.Shutdowner) {
	if err := c.Start(ctx, shutdowner); err != nil {
		log.Printf("Error during crawling: %s", err)
	}
	log.Println("Crawling process finished")
}
