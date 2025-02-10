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

// Params holds the dependencies for creating a new collector
type Params struct {
	fx.In

	BaseURL   string        `name:"baseURL"`
	MaxDepth  int           `name:"maxDepth"`
	RateLimit time.Duration `name:"rateLimit"`
	Debugger  *logger.CollyDebugger
}

// Result holds the dependencies for the collector
type Result struct {
	fx.Out

	Collector *colly.Collector
}

// New creates and returns a new colly collector
func New(p Params) (Result, error) {
	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil {
		return Result{}, fmt.Errorf("failed to parse baseURL: %w", err)
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

	// Declare limitErr outside the if statement to avoid shadowing
	var limitErr error
	if limitErr = collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: CollectorParallelism,
		Delay:       p.RateLimit,
	}); limitErr != nil {
		return Result{}, fmt.Errorf("error setting collector limit: %w", limitErr)
	}

	// Set a timeout for requests
	collector.SetRequestTimeout(RequestTimeout)

	return Result{Collector: collector}, nil
}

// ConfigureLogging configures the logging for the collector
func ConfigureLogging(c *colly.Collector, log logger.Interface) {
	c.OnRequest(func(r *colly.Request) {
		log.Debug("Requesting URL", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Debug("Received response", r.Request.URL.String(), r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Error("Error occurred", r.Request.URL.String(), err)
	})
}

func logRequest(log *logger.CustomLogger, message string, r *colly.Request, data interface{}) {
	log.Info(message, r.URL.String(), r.ID, data)
}

func logVisitError(log *logger.CustomLogger, link string, err error) {
	switch err.Error() {
	case "URL already visited":
		log.Info("URL already visited", link)
	case "Forbidden domain", "Missing URL":
		log.Info(err.Error(), link)
		// case "Max depth limit reached":
		// 	log.Warn("Max depth limit reached", log.Field("link", link))
		// default:
		// 	log.Error("Error visiting link", log.Field("link", link), log.Field("error", err))
	}
}
