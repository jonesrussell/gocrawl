package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
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

// SecurityMiddleware provides security features for the API
type SecurityMiddleware struct {
	cfg             *config.ServerConfig
	log             logger.Interface
	mu              sync.RWMutex
	requests        map[string]map[time.Time]struct{}
	timeProvider    TimeProvider
	rateLimitWindow time.Duration
	cleanupCtx      context.Context
	cleanupCancel   context.CancelFunc
	cleanupWg       sync.WaitGroup
	// rateLimiter tracks request counts per IP
	rateLimiter struct {
		sync.RWMutex
		clients map[string]*rateLimitInfo
		window  time.Duration // Rate limit window
	}
}

// Ensure SecurityMiddleware implements SecurityMiddlewareInterface
var _ SecurityMiddlewareInterface = (*SecurityMiddleware)(nil)

// rateLimitInfo tracks rate limiting information for a client
type rateLimitInfo struct {
	count      int
	lastAccess time.Time
}

// Constants
// No constants needed

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(cfg *config.ServerConfig, log logger.Interface) *SecurityMiddleware {
	ctx, cancel := context.WithCancel(context.Background())
	return &SecurityMiddleware{
		cfg:             cfg,
		log:             log,
		requests:        make(map[string]map[time.Time]struct{}),
		timeProvider:    &realTimeProvider{},
		rateLimitWindow: 5 * time.Second,
		cleanupCtx:      ctx,
		cleanupCancel:   cancel,
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

// Middleware returns the security middleware function
func (m *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Handle CORS preflight requests
		if c.Request.Method == http.MethodOptions {
			if m.cfg.Security.CORS.Enabled {
				origin := c.Request.Header.Get("Origin")
				if m.isOriginAllowed(origin) {
					m.setCORSHeaders(c, origin)
					c.Status(http.StatusNoContent)
					return
				}
				c.Status(http.StatusForbidden)
				return
			}
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		// Handle API key authentication
		if m.cfg.Security.Enabled {
			apiKey := c.GetHeader("X-Api-Key")
			if apiKey != m.cfg.Security.APIKey {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		// Handle rate limiting
		if m.cfg.Security.RateLimit > 0 {
			ip := c.ClientIP()
			if !m.isRequestAllowed(ip) {
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}

		// Set security headers
		m.setSecurityHeaders(c)

		// Handle CORS for non-preflight requests
		if m.cfg.Security.CORS.Enabled {
			origin := c.Request.Header.Get("Origin")
			if m.isOriginAllowed(origin) {
				m.setCORSHeaders(c, origin)
			} else if origin != "" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		c.Next()
	}
}

func (m *SecurityMiddleware) setCORSHeaders(c *gin.Context, origin string) {
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", strings.Join(m.cfg.Security.CORS.AllowedMethods, ", "))
	c.Header("Access-Control-Allow-Headers", strings.Join(m.cfg.Security.CORS.AllowedHeaders, ", "))
	c.Header("Access-Control-Max-Age", strconv.Itoa(m.cfg.Security.CORS.MaxAge))
}

func (m *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range m.cfg.Security.CORS.AllowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// authenticate checks for valid API key
func (m *SecurityMiddleware) authenticate(c *gin.Context) error {
	if m.cfg.Security.APIKey == "" {
		return nil // No API key configured
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return ErrMissingAPIKey
	}

	if apiKey != m.cfg.Security.APIKey {
		return ErrInvalidAPIKey
	}

	return nil
}

// rateLimit implements rate limiting per IP
func (m *SecurityMiddleware) rateLimit(c *gin.Context) error {
	if m.cfg.Security.RateLimit <= 0 {
		return nil // Rate limiting disabled
	}

	clientIP := c.ClientIP()
	now := m.timeProvider.Now()

	m.rateLimiter.Lock()
	defer m.rateLimiter.Unlock()

	info, exists := m.rateLimiter.clients[clientIP]
	if !exists {
		info = &rateLimitInfo{
			count:      1,
			lastAccess: now,
		}
		m.rateLimiter.clients[clientIP] = info
		return nil
	}

	// Reset counter if more than the window has passed
	if now.Sub(info.lastAccess) > m.rateLimiter.window {
		info.count = 1
		info.lastAccess = now
		return nil
	}

	// Check rate limit
	if info.count >= m.cfg.Security.RateLimit {
		return ErrRateLimitExceeded
	}

	info.count++
	info.lastAccess = now
	return nil
}

// handleCORS implements CORS handling
func (m *SecurityMiddleware) handleCORS(c *gin.Context) error {
	if !m.cfg.Security.CORS.Enabled {
		return nil // CORS disabled
	}

	origin := c.GetHeader("Origin")
	if origin == "" {
		return nil // No origin header
	}

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range m.cfg.Security.CORS.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if !allowed {
		return ErrOriginNotAllowed
	}

	// Set CORS headers
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", m.joinStrings(m.cfg.Security.CORS.AllowedMethods))
	c.Header("Access-Control-Allow-Headers", m.joinStrings(m.cfg.Security.CORS.AllowedHeaders))
	c.Header("Access-Control-Max-Age", strconv.Itoa(m.cfg.Security.CORS.MaxAge))

	return nil
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

// joinStrings joins strings with commas
func (m *SecurityMiddleware) joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

// Cleanup periodically removes expired rate limit entries
func (m *SecurityMiddleware) Cleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Second) // Run cleanup every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("Cleanup context cancelled, stopping cleanup routine")
			return
		case <-ticker.C:
			m.mu.Lock()
			now := m.timeProvider.Now()
			window := now.Add(-m.rateLimitWindow)

			// Clean up old requests
			for ip, requests := range m.requests {
				for t := range requests {
					if t.Before(window) {
						delete(requests, t)
					}
				}
				if len(requests) == 0 {
					delete(m.requests, ip)
				}
			}
			m.mu.Unlock()
		}
	}
}

func (m *SecurityMiddleware) isRequestAllowed(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.timeProvider.Now()
	window := now.Add(-m.rateLimitWindow)

	// Clean up old requests
	if m.requests[ip] == nil {
		m.requests[ip] = make(map[time.Time]struct{})
	}

	// Remove expired requests
	for t := range m.requests[ip] {
		if t.Before(window) {
			delete(m.requests[ip], t)
		}
	}

	// Check if rate limit is exceeded
	if len(m.requests[ip]) >= m.cfg.Security.RateLimit {
		return false
	}

	// Add current request
	m.requests[ip][now] = struct{}{}
	return true
}

func (m *SecurityMiddleware) setSecurityHeaders(c *gin.Context) {
	// Set security headers
	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, value := range headers {
		c.Header(header, value)
	}
}

// WaitCleanup waits for cleanup to complete
func (m *SecurityMiddleware) WaitCleanup() {
	m.cleanupWg.Wait()
}
