package collector

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jonesrussell/gocrawl/internal/logger"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"go.uber.org/fx"
)

// Constants for configuration
const (
	Parallelism       = 2 // Maximum parallelism for collector
	RandomDelayFactor = 2 // RandomDelayFactor is used to add randomization to rate limiting
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

// Params holds the parameters for creating a Collector
type Params struct {
	fx.In

	BaseURL   string
	MaxDepth  int
	RateLimit time.Duration
	Debugger  *logger.CollyDebugger
	Logger    logger.Interface
}

// Result holds the collector instance
type Result struct {
	fx.Out

	Collector *colly.Collector
}

// New creates a new collector instance
func New(p Params) (Result, error) {
	// Validate URL
	if p.BaseURL == "" {
		return Result{}, errors.New("base URL cannot be empty")
	}

	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return Result{}, errors.New("invalid base URL: must be a valid HTTP/HTTPS URL")
	}

	// Create collector with base configuration
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(p.MaxDepth),
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RateLimit,
		Parallelism: Parallelism,
	})
	if err != nil {
		return Result{}, errors.New("failed to set rate limit")
	}

	if p.Debugger != nil {
		c.SetDebugger(p.Debugger)
	}

	// Configure logging
	ConfigureLogging(c, p.Logger)

	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
	)

	return Result{Collector: c}, nil
}

// ConfigureLogging sets up logging for the collector
func ConfigureLogging(c *colly.Collector, log logger.Interface) {
	c.OnRequest(func(r *colly.Request) {
		log.Debug("Requesting URL", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Debug("Received response",
			"url", r.Request.URL.String(),
			"status", r.StatusCode,
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Error("Error occurred",
			"url", r.Request.URL.String(),
			"error", err.Error(),
		)
	})
}
