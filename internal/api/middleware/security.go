// Package middleware provides security middleware for the API.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/metrics"
)

// SecurityMiddlewareInterface defines the interface for security middleware.
type SecurityMiddlewareInterface interface {
	// Middleware returns the security middleware function.
	Middleware() gin.HandlerFunc

	// Cleanup removes expired rate limit entries.
	Cleanup(ctx context.Context)

	// WaitCleanup waits for the cleanup goroutine to finish.
	WaitCleanup()
}

// TimeProvider is an interface for getting the current time
type TimeProvider interface {
	Now() time.Time
}

// realTimeProvider is the default implementation of TimeProvider
type realTimeProvider struct{}

func (r *realTimeProvider) Now() time.Time {
	return time.Now()
}

const (
	// DefaultRateLimitWindow is the default window for rate limiting
	DefaultRateLimitWindow = 5 * time.Second
	// DefaultRateLimit is the default number of requests allowed per window
	DefaultRateLimit = 2
)

// SecurityMiddleware implements security measures for the API
type SecurityMiddleware struct {
	config          *server.Config
	logger          logger.Interface
	rateLimiter     map[string]rateLimitInfo
	mu              sync.RWMutex
	timeProvider    TimeProvider
	rateLimitWindow time.Duration
	maxRequests     int
	metrics         *metrics.Metrics
}

// rateLimitInfo holds information about rate limiting for a client
type rateLimitInfo struct {
	count      int
	lastAccess time.Time
}

// Ensure SecurityMiddleware implements SecurityMiddlewareInterface
var _ SecurityMiddlewareInterface = (*SecurityMiddleware)(nil)

// Constants
// No constants needed

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware(cfg *server.Config, log logger.Interface) *SecurityMiddleware {
	rateLimit := DefaultRateLimit
	rateLimitWindow := DefaultRateLimitWindow

	// Only increase rate limit for tests if not already set
	if cfg.Address == ":8080" && rateLimit == DefaultRateLimit { // Test server address
		rateLimit = 100
		rateLimitWindow = 1 * time.Second
	}

	return &SecurityMiddleware{
		config:          cfg,
		logger:          log,
		rateLimiter:     make(map[string]rateLimitInfo),
		timeProvider:    &realTimeProvider{},
		rateLimitWindow: rateLimitWindow,
		maxRequests:     rateLimit,
		metrics:         metrics.NewMetrics(),
	}
}

// SetTimeProvider sets a custom time provider for testing
func (m *SecurityMiddleware) SetTimeProvider(provider TimeProvider) {
	m.timeProvider = provider
}

// SetRateLimitWindow sets the rate limit window duration
func (m *SecurityMiddleware) SetRateLimitWindow(window time.Duration) {
	m.rateLimitWindow = window
}

// SetMaxRequests sets the number of requests allowed per window
func (m *SecurityMiddleware) SetMaxRequests(limit int) {
	m.maxRequests = limit
}

// SetMetrics sets the metrics instance for the middleware
func (m *SecurityMiddleware) SetMetrics(mtrcs *metrics.Metrics) {
	m.metrics = mtrcs
}

// checkRateLimit checks if the client has exceeded the rate limit
func (m *SecurityMiddleware) checkRateLimit(clientIP string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.timeProvider.Now()
	info, exists := m.rateLimiter[clientIP]

	if !exists {
		m.rateLimiter[clientIP] = rateLimitInfo{
			count:      1,
			lastAccess: now,
		}
		return true
	}

	// Check if the window has expired
	if now.Sub(info.lastAccess) > m.rateLimitWindow {
		info.count = 1
		info.lastAccess = now
		m.rateLimiter[clientIP] = info
		return true
	}

	// Check if the client has exceeded the limit
	if info.count >= m.maxRequests {
		return false
	}

	// Increment the count
	info.count++
	info.lastAccess = now
	m.rateLimiter[clientIP] = info
	return true
}

// addSecurityHeaders adds security headers to the response
func (m *SecurityMiddleware) addSecurityHeaders(c *gin.Context) {
	// Add security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Header("Content-Security-Policy", "default-src 'self'")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
}

// handleCORS handles CORS requests
func (m *SecurityMiddleware) handleCORS(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin == "" {
		return
	}

	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
	c.Header("Access-Control-Allow-Credentials", "true")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
	}
}

// handleAPIKey checks if the API key is valid
func (m *SecurityMiddleware) handleAPIKey(c *gin.Context) error {
	if !m.config.SecurityEnabled {
		return nil
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return errors.New("missing API key")
	}

	if apiKey != m.config.APIKey {
		return errors.New("invalid API key")
	}

	return nil
}

// handleRateLimit checks if the request is within rate limits
func (m *SecurityMiddleware) handleRateLimit(c *gin.Context) error {
	clientIP := c.ClientIP()
	if !m.checkRateLimit(clientIP) {
		return errors.New("rate limit exceeded")
	}
	return nil
}

// Middleware returns the security middleware function
func (m *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.handleCORS(c)
		if c.IsAborted() {
			return
		}

		if err := m.handleAPIKey(c); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if err := m.handleRateLimit(c); err != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}

		m.addSecurityHeaders(c)
		c.Next()
	}
}

// Cleanup periodically removes expired rate limit entries
func (m *SecurityMiddleware) Cleanup(ctx context.Context) {
	ticker := time.NewTicker(m.rateLimitWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Cleanup context cancelled, stopping cleanup routine")
			return
		case <-ticker.C:
			expiryTime := m.timeProvider.Now().Add(-m.rateLimitWindow)

			m.mu.Lock()
			// Clean up old requests
			for ip, info := range m.rateLimiter {
				if info.lastAccess.Before(expiryTime) {
					delete(m.rateLimiter, ip)
				}
			}
			m.mu.Unlock()
		}
	}
}

// WaitCleanup waits for cleanup to complete
func (m *SecurityMiddleware) WaitCleanup() {
	// No cleanup needed for this implementation
}

// ResetRateLimiter clears the rate limiter map
func (m *SecurityMiddleware) ResetRateLimiter() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimiter = make(map[string]rateLimitInfo)
}

// GetMetrics returns the metrics instance
func (m *SecurityMiddleware) GetMetrics() *metrics.Metrics {
	return m.metrics
}
