package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// TimeProvider allows mocking time in tests
type TimeProvider interface {
	Now() time.Time
}

// realTimeProvider provides actual time
type realTimeProvider struct{}

func (r *realTimeProvider) Now() time.Time {
	return time.Now()
}

// SecurityMiddleware provides security features for the API
type SecurityMiddleware struct {
	config *config.ServerConfig
	logger logger.Interface
	time   TimeProvider
	// rateLimiter tracks request counts per IP
	rateLimiter struct {
		sync.RWMutex
		clients map[string]*rateLimitInfo
		window  time.Duration // Rate limit window
	}
}

// rateLimitInfo tracks rate limiting information for a client
type rateLimitInfo struct {
	count      int
	lastAccess time.Time
}

// NewSecurityMiddleware creates a new security middleware instance
func NewSecurityMiddleware(cfg *config.ServerConfig, logger logger.Interface) *SecurityMiddleware {
	m := &SecurityMiddleware{
		config: cfg,
		logger: logger,
		time:   &realTimeProvider{},
	}
	m.rateLimiter.clients = make(map[string]*rateLimitInfo)
	m.rateLimiter.window = time.Minute // Default to 1 minute
	return m
}

// SetTimeProvider sets a custom time provider for testing
func (m *SecurityMiddleware) SetTimeProvider(provider TimeProvider) {
	m.time = provider
}

// SetRateLimitWindow sets the rate limit window for testing
func (m *SecurityMiddleware) SetRateLimitWindow(window time.Duration) {
	m.rateLimiter.Lock()
	defer m.rateLimiter.Unlock()
	m.rateLimiter.window = window
}

// Middleware returns the security middleware function
func (m *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip security checks if disabled
		if !m.config.Security.Enabled {
			c.Next()
			return
		}

		// Apply security features in order
		if err := m.authenticate(c); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if err := m.rateLimit(c); err != nil {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		if err := m.handleCORS(c); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CORS error"})
			return
		}

		// Add security headers
		m.addSecurityHeaders(c)

		c.Next()
	}
}

// authenticate checks for valid API key
func (m *SecurityMiddleware) authenticate(c *gin.Context) error {
	if m.config.Security.APIKey == "" {
		return nil // No API key configured
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return ErrMissingAPIKey
	}

	if apiKey != m.config.Security.APIKey {
		return ErrInvalidAPIKey
	}

	return nil
}

// rateLimit implements rate limiting per IP
func (m *SecurityMiddleware) rateLimit(c *gin.Context) error {
	if m.config.Security.RateLimit <= 0 {
		return nil // Rate limiting disabled
	}

	clientIP := c.ClientIP()
	now := m.time.Now()

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
	if info.count >= m.config.Security.RateLimit {
		return ErrRateLimitExceeded
	}

	info.count++
	info.lastAccess = now
	return nil
}

// handleCORS implements CORS handling
func (m *SecurityMiddleware) handleCORS(c *gin.Context) error {
	if !m.config.Security.CORS.Enabled {
		return nil // CORS disabled
	}

	origin := c.GetHeader("Origin")
	if origin == "" {
		return nil // No origin header
	}

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range m.config.Security.CORS.AllowedOrigins {
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
	c.Header("Access-Control-Allow-Methods", m.joinStrings(m.config.Security.CORS.AllowedMethods))
	c.Header("Access-Control-Allow-Headers", m.joinStrings(m.config.Security.CORS.AllowedHeaders))
	c.Header("Access-Control-Max-Age", strconv.Itoa(m.config.Security.CORS.MaxAge))

	// Handle preflight requests
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
	}

	return nil
}

// addSecurityHeaders adds security-related headers to the response
func (m *SecurityMiddleware) addSecurityHeaders(c *gin.Context) {
	// Prevent clickjacking
	c.Header("X-Frame-Options", "DENY")
	// Enable XSS protection
	c.Header("X-XSS-Protection", "1; mode=block")
	// Prevent MIME type sniffing
	c.Header("X-Content-Type-Options", "nosniff")
	// Enable HSTS
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	// Referrer policy
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	// Content security policy
	c.Header("Content-Security-Policy", "default-src 'self'")
}

// joinStrings joins a slice of strings with commas
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

// Cleanup removes expired rate limit entries
func (m *SecurityMiddleware) Cleanup(ctx context.Context) {
	ticker := time.NewTicker(m.rateLimiter.window)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.rateLimiter.Lock()
			now := m.time.Now()
			for ip, info := range m.rateLimiter.clients {
				if now.Sub(info.lastAccess) > m.rateLimiter.window {
					delete(m.rateLimiter.clients, ip)
				}
			}
			m.rateLimiter.Unlock()
		}
	}
}

// Error definitions
var (
	ErrMissingAPIKey     = errors.New("missing API key")
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrOriginNotAllowed  = errors.New("origin not allowed")
)
