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

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	app := createApp(*urlPtr, *maxDepthPtr, *rateLimitPtr, cfg)

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

func createApp(url string, maxDepth int, rateLimit time.Duration, cfg *config.Config) *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		fx.Provide(
			func(lc fx.Lifecycle, loggerInstance *logger.CustomLogger) (*crawler.Crawler, error) {
				return initializeCrawler(url, maxDepth, rateLimit, loggerInstance, cfg)
			},
		),
		fx.WithLogger(func(loggerInstance *logger.CustomLogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: loggerInstance.GetZapLogger()}
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
		close(done)
	}()

	<-done
	log.Println("Crawling process finished")

	// Wait for all requests to complete
	c.Collector.Wait()

	// Trigger shutdown
	if err := shutdowner.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

func initializeCrawler(url string, maxDepth int, rateLimit time.Duration, loggerInstance *logger.CustomLogger, cfg *config.Config) (*crawler.Crawler, error) {
	debugger := logger.NewCustomDebugger(loggerInstance)
	return crawler.NewCrawler(url, maxDepth, rateLimit, debugger, loggerInstance, cfg)
}
