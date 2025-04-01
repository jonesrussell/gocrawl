// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

import (
	"fmt"
	"net/http"
	"net/url"
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
	// maxRetries is the maximum number of retry attempts for failed requests.
	maxRetries = 2
	// maxConnections is the maximum number of concurrent connections.
	maxConnections = 10
	// idleTimeoutSeconds is the timeout for idle connections.
	idleTimeoutSeconds = 30
	// tlsTimeoutSeconds is the timeout for TLS handshake.
	tlsTimeoutSeconds = 10
	// expectContinueTimeoutSeconds is the timeout for expect-continue handshake.
	expectContinueTimeoutSeconds = 1
	// backoffMultiplier is the multiplier for exponential backoff.
	backoffMultiplier = 2
)

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
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

// handleRetry handles retry logic for failed requests
func (s *Setup) handleRetry(
	c *colly.Collector,
	r *colly.Response,
	err error,
	url string,
	count int,
	retryCount map[string]int,
) {
	retryCount[url] = count + 1
	s.config.Logger.Warn("Retrying request",
		"url", url,
		"attempt", count+1,
		"max_attempts", maxRetries,
		"status_code", getStatusCode(r),
		"error", err)

	// Add exponential backoff
	backoff := time.Duration(count+1) * backoffMultiplier * time.Second
	time.Sleep(backoff)

	var retryErr error
	if r != nil {
		retryErr = r.Request.Retry()
	} else {
		retryErr = c.Visit(url)
	}

	if retryErr != nil {
		s.config.Logger.Error("Retry failed",
			"url", url,
			"attempt", count+1,
			"error", retryErr)
	}
}

// extractURLFromError attempts to extract a URL from an error message
func extractURLFromError(err error, domain string, logger Logger) (string, bool) {
	errStr := err.Error()
	start := strings.Index(errStr, "\"")
	end := strings.LastIndex(errStr, "\"")
	if start != -1 && end != -1 && end > start {
		return errStr[start+1 : end], true
	}
	logger.Error("Request failed with nil response",
		"error", err,
		"domain", domain)
	return "", false
}

// shouldRetry determines if a request should be retried
func shouldRetry(r *colly.Response, err error, count int) bool {
	return (strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "temporary") ||
		strings.Contains(err.Error(), "TLS handshake") ||
		strings.Contains(err.Error(), "connection refused") ||
		(r != nil && r.StatusCode >= 500 && r.StatusCode < 600)) && count < maxRetries
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

	// Create collector with domain restrictions
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(s.config.MaxDepth),
		colly.MaxBodySize(maxBodySizeMB*megabyte),
		colly.AllowedDomains(allowedDomains...),
	)

	c.DisableCookies()

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
	c.SetRequestTimeout(time.Duration(requestTimeoutSeconds) * time.Second)

	retryCount := make(map[string]int)

	c.OnError(func(r *colly.Response, err error) {
		var u string
		if r != nil {
			u = r.Request.URL.String()
		} else if err != nil {
			var ok bool
			if u, ok = extractURLFromError(err, domain, s.config.Logger); !ok {
				return
			}
		}

		count := retryCount[u]
		if shouldRetry(r, err, count) {
			s.handleRetry(c, r, err, u, count, retryCount)
		} else {
			s.config.Logger.Error("Request error",
				"u", u,
				"status_code", getStatusCode(r),
				"error", err,
				"retries", count,
				"max_retries", maxRetries)
		}
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
	// Configure rate limiting with domain-specific rules
	if err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: s.config.RandomDelay,
		Parallelism: s.config.Parallelism,
	}); err != nil {
		return fmt.Errorf("failed to set rate limit: %w", err)
	}

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
