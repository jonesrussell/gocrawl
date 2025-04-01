// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	// kilobyte represents 1024 bytes.
	kilobyte = 1024
	// megabyte represents 1024 kilobytes.
	megabyte = 1024 * kilobyte
	// maxBodySizeMB is the maximum size in megabytes for response bodies.
	maxBodySizeMB = 10
	// requestTimeoutSeconds is the number of seconds to wait for a response.
	requestTimeoutSeconds = 45
	// maxConnections is the maximum number of concurrent connections.
	maxConnections = 10
	// idleTimeoutSeconds is the timeout for idle connections.
	idleTimeoutSeconds = 30
	// tlsTimeoutSeconds is the timeout for TLS handshake.
	tlsTimeoutSeconds = 10
	// expectContinueTimeoutSeconds is the timeout for expect-continue handshake.
	expectContinueTimeoutSeconds = 1
)

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
}

// Setup handles the setup and configuration of the collector.
// It manages the creation and configuration of the Colly collector instance,
// including rate limiting, debugging, logging, and content processing.
type Setup struct {
	// config contains the collector configuration
	config *Config
}

// NewSetup creates a new collector setup instance.
// It initializes the setup with the provided configuration.
//
// Parameters:
//   - config: The collector configuration
//
// Returns:
//   - *Setup: The initialized setup instance
func NewSetup(config *Config) *Setup {
	return &Setup{
		config: config,
	}
}

// CreateBaseCollector creates a new collector with base configuration.
// It sets up the collector with basic settings including:
// - Asynchronous operation
// - Maximum depth limit
// - Domain restrictions
// - Request timeouts and retries
//
// Parameters:
//   - domain: The domain to restrict crawling to
//
// Returns:
//   - *colly.Collector: The configured collector instance
func (s *Setup) CreateBaseCollector(domain string) *colly.Collector {
	// Parse the base URL to get the domain
	baseURL, err := url.Parse(s.config.BaseURL)
	if err != nil {
		s.config.Logger.Error("Failed to parse base URL", "error", err)
		return nil
	}

	// Set allowed domains to the base URL's domain
	allowedDomains := []string{baseURL.Hostname()}

	// Create collector with domain restrictions and duplicate handling
	c := colly.NewCollector(
		colly.Async(false), // Disable async to ensure proper rate limiting
		colly.MaxDepth(s.config.MaxDepth),
		colly.MaxBodySize(maxBodySizeMB*megabyte),
		colly.AllowedDomains(allowedDomains...),
		colly.URLFilters(
			regexp.MustCompile(`.*`), // Allow all URLs within allowed domains
		),
	)

	// Configure duplicate handling
	c.SetRequestTimeout(time.Duration(requestTimeoutSeconds) * time.Second)
	c.DisableCookies()

	// Configure transport with keep-alives disabled to prevent connection reuse
	transport := &http.Transport{
		DisableKeepAlives:     true,
		MaxIdleConns:          maxConnections,
		MaxIdleConnsPerHost:   maxConnections,
		MaxConnsPerHost:       maxConnections,
		IdleConnTimeout:       time.Duration(idleTimeoutSeconds) * time.Second,
		TLSHandshakeTimeout:   time.Duration(tlsTimeoutSeconds) * time.Second,
		ExpectContinueTimeout: time.Duration(expectContinueTimeoutSeconds) * time.Second,
	}

	c.WithTransport(transport)

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		s.config.Logger.Error("Request error",
			"url", r.Request.URL.String(),
			"status_code", getStatusCode(r),
			"error", err)
	})

	return c
}

// ConfigureCollector sets up the collector with all necessary settings.
// It configures:
// - Rate limiting
// - Debugging
// - Logging
// - Content processing
//
// Parameters:
//   - c: The collector instance to configure
//
// Returns:
//   - error: Any error that occurred during configuration
func (s *Setup) ConfigureCollector(c *colly.Collector) error {
	// Parse rate limit duration
	rateLimit, err := time.ParseDuration(s.config.RateLimit)
	if err != nil {
		return fmt.Errorf("failed to parse rate limit: %w", err)
	}

	// Log rate limiting configuration
	s.config.Logger.Debug("Configuring rate limiting",
		"rate_limit", rateLimit,
		"random_delay", rateLimit/2,
		"parallelism", s.config.Parallelism,
		"max_depth", s.config.MaxDepth,
		"source", s.config.Source.Name,
		"base_url", s.config.Source.URL)

	// Configure rate limiting with domain-specific rules
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       rateLimit,
		RandomDelay: rateLimit / 2, // Add some randomization to avoid thundering herd
		Parallelism: s.config.Parallelism,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

	// Add detailed request timing and depth logging
	c.OnRequest(func(r *colly.Request) {
		s.config.Logger.Debug("Starting request",
			"url", r.URL.String(),
			"domain", r.URL.Hostname(),
			"depth", r.Depth,
			"max_depth", s.config.MaxDepth,
			"rate_limit", rateLimit.String(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
			"source", s.config.Source.Name)
	})

	// Add response timing logging with more details
	c.OnResponse(func(r *colly.Response) {
		s.config.Logger.Debug("Completed request",
			"url", r.Request.URL.String(),
			"domain", r.Request.URL.Hostname(),
			"depth", r.Request.Depth,
			"status", r.StatusCode,
			"timestamp", time.Now().Format(time.RFC3339Nano),
			"source", s.config.Source.Name)
	})

	// Add error logging with more context
	c.OnError(func(r *colly.Response, err error) {
		s.config.Logger.Error("Request failed",
			"url", r.Request.URL.String(),
			"domain", r.Request.URL.Hostname(),
			"depth", r.Request.Depth,
			"status", r.StatusCode,
			"error", err.Error(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
			"source", s.config.Source.Name)
	})

	// Configure debugger and logging if enabled
	if s.config.Debugger != nil {
		c.SetDebugger(s.config.Debugger)
	}
	ConfigureLogging(c, s.config.Logger)

	// Configure content processing with provided processors
	contentParams := ContentParams{
		Logger:           s.config.Logger,
		Source:           s.config.Source,
		ArticleProcessor: s.config.ArticleProcessor,
		ContentProcessor: s.config.ContentProcessor,
	}
	configureContentProcessing(c, contentParams)

	return nil
}

// ValidateURL validates the base URL.
// It ensures the URL is:
// - Well-formed
// - Uses HTTP or HTTPS protocol
//
// Returns:
//   - error: Any validation error that occurred
func (s *Setup) ValidateURL() error {
	parsedURL, err := url.Parse(s.config.BaseURL)
	if err != nil || (!strings.HasPrefix(parsedURL.Scheme, "http") && !strings.HasPrefix(parsedURL.Scheme, "https")) {
		return fmt.Errorf("invalid base URL: %s, must be a valid HTTP/HTTPS URL", s.config.BaseURL)
	}
	return nil
}

// getStatusCode safely returns the status code from a response
func getStatusCode(r *colly.Response) int {
	if r != nil {
		return r.StatusCode
	}
	return 0
}
