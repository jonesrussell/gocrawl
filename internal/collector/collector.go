package collector

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/fx"
)

// Constants for configuration
const (
	RequestTimeout       = 30 * time.Second // Timeout for requests
	CollectorParallelism = 2                // Maximum parallelism for collector
	RandomDelayFactor    = 2                // RandomDelayFactor is used to add randomization to rate limiting
)

// DebuggerInterface is an interface for the debugger
type DebuggerInterface interface {
	Init() error
	OnRequest(e *colly.Request)
	OnResponse(e *colly.Response)
	OnError(e *colly.Response, err error)
	OnEvent(e *debug.Event)
	Event(e *debug.Event)
}

// Params holds the dependencies for creating a new collector
type Params struct {
	fx.In

	BaseURL   string        `name:"baseURL"`
	MaxDepth  int           `name:"maxDepth"`
	RateLimit time.Duration `name:"rateLimit"`
	Debugger  DebuggerInterface
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
		colly.AllowedDomains(allowedDomain),
		colly.URLFilters(
			regexp.MustCompile(fmt.Sprintf("^https?://%s/.*", allowedDomain)),
		),
		colly.ParseHTTPErrorResponse(),
	)

	// Add URL normalization
	collector.OnRequest(func(r *colly.Request) {
		r.URL.RawQuery = "" // Remove query parameters
		if !strings.HasPrefix(r.URL.Scheme, "http") {
			r.URL.Scheme = "https"
		}
	})

	// Configure limits
	err = collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: CollectorParallelism,
		Delay:       p.RateLimit,
		RandomDelay: p.RateLimit / RandomDelayFactor, // Add some randomization to be more polite
	})
	if err != nil {
		return Result{}, fmt.Errorf("error setting collector limit: %w", err)
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
