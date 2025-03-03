package collector

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

// New creates a new collector instance
func New(p Params) (Result, error) {
	if err := ValidateParams(p); err != nil {
		return Result{}, err
	}

	parsedURL, err := url.Parse(p.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return Result{}, fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", p.BaseURL)
	}

	// Extract the domain from the BaseURL
	domain := parsedURL.Hostname()

	// Create collector with base configuration
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(p.MaxDepth),
		colly.AllowedDomains(domain),
	)

	// Set rate limiting
	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: p.RandomDelay,
		Parallelism: p.Parallelism,
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to set rate limit: %w", err)
	}

	if p.Debugger != nil {
		c.SetDebugger(p.Debugger)
	}

	// Configure logging
	ConfigureLogging(c, p.Logger)

	// Configure content processing
	configureContentProcessing(c, p)

	p.Logger.Debug("Collector created",
		"baseURL", p.BaseURL,
		"maxDepth", p.MaxDepth,
		"rateLimit", p.RateLimit,
		"parallelism", p.Parallelism,
	)

	return Result{Collector: c}, nil
}
