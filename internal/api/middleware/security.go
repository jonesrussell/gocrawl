package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/types"
)

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
)

// SecurityMiddleware implements security measures for the API
type SecurityMiddleware struct {
	config          *config.ServerConfig
	logger          logger.Interface
	rateLimiter     map[string]rateLimitInfo
	mu              sync.RWMutex
	timeProvider    TimeProvider
	rateLimitWindow time.Duration
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
func NewSecurityMiddleware(cfg *config.ServerConfig, log logger.Interface) *SecurityMiddleware {
	return &SecurityMiddleware{
		config:          cfg,
		logger:          log,
		rateLimiter:     make(map[string]rateLimitInfo),
		timeProvider:    &realTimeProvider{},
		rateLimitWindow: DefaultRateLimitWindow,
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

// Middleware returns a Gin middleware that applies security measures
func (m *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply security headers first
		m.addSecurityHeaders(c)

		// Handle CORS
		if err := m.handleCORS(c); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIError{
				Code:    http.StatusForbidden,
				Message: err.Error(),
			})
			return
		}

		// Authenticate request first
		if err := m.authenticate(c); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
			return
		}

		// Apply rate limiting after authentication
		if err := m.rateLimit(c); err != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, types.APIError{
				Code:    http.StatusTooManyRequests,
				Message: err.Error(),
			})
			return
		}

		c.Next()
	}
}

// addSecurityHeaders adds security-related headers to the response
func (m *SecurityMiddleware) addSecurityHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	c.Header("Content-Security-Policy", "default-src 'self'")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
}

// handleCORS handles CORS headers and preflight requests
func (m *SecurityMiddleware) handleCORS(c *gin.Context) error {
	origin := c.GetHeader("Origin")
	if origin == "" {
		return nil
	}

	// Check if origin is allowed
	if !m.isAllowedOrigin(origin) {
		return fmt.Errorf("origin not allowed: %s", origin)
	}

	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Api-Key")
	c.Header("Access-Control-Max-Age", "86400") // 24 hours

	// Handle preflight requests
	if c.Request.Method == http.MethodOptions {
		c.Header("Access-Control-Allow-Credentials", "true")
		c.AbortWithStatus(http.StatusNoContent) // Use 204 for preflight
		return nil
	}

	return nil
}

// isAllowedOrigin checks if the given origin is allowed
func (m *SecurityMiddleware) isAllowedOrigin(origin string) bool {
	for _, allowed := range m.config.Security.CORS.AllowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}

// rateLimit applies rate limiting to the request
func (m *SecurityMiddleware) rateLimit(c *gin.Context) error {
	ip := c.ClientIP()
	now := m.timeProvider.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	info, exists := m.rateLimiter[ip]
	if !exists {
		info = rateLimitInfo{
			count:      1,
			lastAccess: now,
		}
		m.rateLimiter[ip] = info
		return nil
	}

	if now.Sub(info.lastAccess) > m.rateLimitWindow {
		info.count = 1
		info.lastAccess = now
	} else if info.count >= m.config.Security.RateLimit {
		return errors.New("rate limit exceeded")
	} else {
		info.count++
		info.lastAccess = now
	}
	m.rateLimiter[ip] = info
	return nil
}

// authenticate authenticates the request using API key
func (m *SecurityMiddleware) authenticate(c *gin.Context) error {
	apiKey := c.GetHeader("X-Api-Key")
	if apiKey == "" {
		return errors.New("missing API key")
	}

	if apiKey != m.config.Security.APIKey {
		return errors.New("invalid API key")
	}

	return nil
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
