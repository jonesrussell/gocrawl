package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
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
	// Define command-line flags
	urlPtr := flag.String("url", "http://example.com", "The URL to crawl")
	maxDepthPtr := flag.Int("maxDepth", 2, "The maximum depth to crawl")
	rateLimitPtr := flag.Duration("rateLimit", 5*time.Second, "Rate limit between requests")
	flag.Parse() // Parse the command-line flags

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	// Create and run the Fx application
	app := fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		fx.Provide(
			func(lc fx.Lifecycle, loggerInstance *logger.CustomLogger) (*crawler.Crawler, error) {
				return initializeCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, loggerInstance, cfg)
			},
		),
		fx.WithLogger(func(loggerInstance *logger.CustomLogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: loggerInstance.GetZapLogger()} // Use the underlying zap logger
		}),
		fx.Invoke(func(lc fx.Lifecycle, c *crawler.Crawler) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
						defer cancel()
						c.Start(ctx, *urlPtr) // Start the crawler

						<-c.Done // Wait for crawling to complete
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Shutdown process initiated...") // Log when shutdown starts
					logGoroutineCount()                          // Log active goroutines before waiting
					// Wait for the crawling to finish
					select {
					case <-c.Done: // Wait for the Done channel to be closed
						log.Println("Crawling completed, shutting down...")
					case <-ctx.Done(): // Handle context timeout
						log.Println("Shutdown timeout, forcing exit...")
					}
					c.Collector.Wait()                             // Wait for all requests to finish
					logGoroutineCount()                            // Log active goroutines after waiting
					log.Println("Crawling completed successfully") // Log when crawling is done
					return nil
				},
			})
		}),
	)

	// Handle shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		app.Stop(context.Background())
	}()

	app.Run()
}

func initializeCrawler(url string, maxDepth int, rateLimit time.Duration, loggerInstance *logger.CustomLogger, cfg *config.Config) (*crawler.Crawler, error) {
	// Create a new Debugger
	debugger := logger.NewCustomDebugger(loggerInstance)

	// Create a new Crawler with the config
	crawlerInstance, err := crawler.NewCrawler(url, maxDepth, rateLimit, debugger, loggerInstance, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize crawler: %w", err)
	}

	return crawlerInstance, nil
}

func logGoroutineCount() {
	numGoroutines := runtime.NumGoroutine()
	log.Printf("Number of active goroutines: %d\n", numGoroutines)
}
