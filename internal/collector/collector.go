package collector

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"

	"github.com/gocolly/colly/v2"
	"go.uber.org/fx"
)

// Constants for configuration
const (
	RequestTimeout       = 30 * time.Second // Timeout for requests
	CollectorParallelism = 2                // Maximum parallelism for collector
)

// CollectorParams holds the dependencies for creating a new collector
type CollectorParams struct {
	fx.In

	BaseURL   string        `name:"baseURL"`
	MaxDepth  int           `name:"maxDepth"`
	RateLimit time.Duration `name:"rateLimit"`
	Debugger  *logger.CustomDebugger
}

// New creates and returns a new colly collector
func New(p CollectorParams) (*colly.Collector, error) {
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL: %w", err)
	}
	allowedDomain := parsedURL.Hostname()

	collector := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(p.MaxDepth),
		colly.Debugger(p.Debugger),
		colly.AllowedDomains(
			allowedDomain,
			"http://"+allowedDomain,
			"https://"+allowedDomain,
		),
	)

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: CollectorParallelism,
		Delay:       p.RateLimit,
	})

	// Set a timeout for requests
	collector.SetRequestTimeout(RequestTimeout)

	return collector, nil
}

// ConfigureLogging sets up logging for the collector
func ConfigureLogging(collector *colly.Collector, log *logger.CustomLogger) {
	collector.OnRequest(func(r *colly.Request) {
		startTime := time.Now()
		logRequest(log, "Requesting URL", r, startTime)

		defer func() {
			duration := time.Since(startTime)
			logRequest(log, "Request completed", r, duration)
		}()
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if err := e.Request.Visit(link); err != nil {
			logVisitError(log, link, err)
		}
	})
}

func logRequest(log *logger.CustomLogger, message string, r *colly.Request, data interface{}) {
	log.Info(message, log.Field("url", r.URL.String()), log.Field("request_id", r.ID), log.Field("data", data))
}

func logVisitError(log *logger.CustomLogger, link string, err error) {
	switch err.Error() {
	case "URL already visited":
		log.Info("URL already visited", log.Field("link", link))
	case "Forbidden domain", "Missing URL":
		log.Info(err.Error(), log.Field("link", link))
		// case "Max depth limit reached":
		// 	log.Warn("Max depth limit reached", log.Field("link", link))
		// default:
		// 	log.Error("Error visiting link", log.Field("link", link), log.Field("error", err))
	}
}
