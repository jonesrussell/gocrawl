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

	// Handle shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		app.Stop(context.Background())
	}()

	app.Run()
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
	shutdowner.Shutdown()

	log.Println("Crawling process finished")
}

func initializeCrawler(url string, maxDepth int, rateLimit time.Duration, loggerInstance *logger.CustomLogger, cfg *config.Config) (*crawler.Crawler, error) {
	debugger := logger.NewCustomDebugger(loggerInstance)
	return crawler.NewCrawler(url, maxDepth, rateLimit, debugger, loggerInstance, cfg)
}
