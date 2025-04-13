// Package middleware provides security middleware for the API.
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
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
