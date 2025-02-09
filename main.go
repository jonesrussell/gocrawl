package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Run the application in a goroutine
	go func() {
		if err := app.Start(context.Background()); err != nil {
			log.Fatalf("Application error: %v", err)
		}
	}()

	// Wait for a signal
	<-sigs
	log.Println("Received shutdown signal, exiting...")
	// No need to call app.Shutdown() as fx handles it automatically
}

func createApp(url string, maxDepth int, rateLimit time.Duration) *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
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
			crawler.NewCrawler,
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
			go startCrawling(ctx, c, shutdowner)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutdown process initiated...")
			return nil
		},
	})
}

func startCrawling(ctx context.Context, c *crawler.Crawler, shutdowner fx.Shutdowner) {
	done := make(chan struct{})

	ctx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()

	log.Println("Starting the crawling process")

	go func() {
		err := c.Start(ctx)
		if err != nil {
			log.Printf("Error during crawling: %s", err)
		}
		log.Println("Crawler.Start() completed, closing done channel")
		close(done)
	}()

	log.Println("Waiting for done signal...")
	<-done
	log.Println("Crawling process finished")

	log.Println("Initiating shutdown...")
	if err := shutdowner.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Shutdown completed")
}
