package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// logPrinter is a struct that implements fx.Printer
type logPrinter struct{}

func (lp *logPrinter) Printf(format string, args ...any) {
	log.Printf(format, args...) // Use log.Printf for formatted logging
}

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

	// Initialize the logger based on the environment
	loggerInstance := newLogger(cfg)

	// Print configuration if APP_DEBUG is true
	if cfg.AppDebug {
		loggerInstance.Info("Crawling Configuration",
			loggerInstance.Field("url", *urlPtr),
			loggerInstance.Field("maxDepth", *maxDepthPtr),
			loggerInstance.Field("rateLimit", rateLimitPtr.String()))
	}

	// Create and run the Fx application
	app := fx.New(
		fx.Provide(
			// Provide the logger
			func() *logger.CustomLogger {
				return loggerInstance
			},
			func() fx.Printer {
				return &logPrinter{} // Provide the custom fx.Printer
			},
			func(lc fx.Lifecycle) (*crawler.Crawler, error) {
				return initializeCrawler(*urlPtr, *maxDepthPtr, *rateLimitPtr, loggerInstance, cfg)
			},
		),
		fx.WithLogger(func(l fx.Printer) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: loggerInstance.GetZapLogger()} // Use the underlying zap logger
		}),
		fx.Invoke(func(lc fx.Lifecycle, c *crawler.Crawler) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
						defer cancel()
						c.Start(ctx, *urlPtr) // Pass the URL pointer as the second argument

						<-c.Done // Wait for crawling to complete
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					c.Collector.Wait() // Wait for all requests to finish
					return nil
				},
			})
		}),
	)

	app.Run()
}

func newLogger(cfg *config.Config) *logger.CustomLogger {
	if cfg.AppEnv == "development" {
		loggerInstance, err := logger.NewDevelopmentLogger() // Use development logger
		if err != nil {
			log.Fatalf("Error initializing development logger: %s", err)
		}
		return loggerInstance
	}

	loggerInstance, err := logger.NewCustomLogger() // Use production logger
	if err != nil {
		log.Fatalf("Error initializing production logger: %s", err)
	}
	return loggerInstance
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
