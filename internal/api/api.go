// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// Constants
const (
	readHeaderTimeout = 10 * time.Second // Timeout for reading headers
	shutdownTimeout   = 5 * time.Second  // Timeout for graceful shutdown
)

// SetupRouter creates and configures the Gin router with all routes
func SetupRouter(
	log logger.Interface,
	searchManager SearchManager,
	cfg config.Interface,
) (*gin.Engine, *middleware.SecurityMiddleware) {
	// Disable Gin's default logging
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(log))

	// Create security middleware
	security := middleware.NewSecurityMiddleware(cfg.GetServerConfig(), log)
	router.Use(security.Middleware())

	// Define routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/search", handleSearch(searchManager))

	return router, security
}

// loggingMiddleware creates a middleware that logs request details
func loggingMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		param := gin.LogFormatterParams{
			Path:         path,
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			Latency:      time.Since(start),
			ClientIP:     c.ClientIP(),
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}

		log.Info("Gin request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"query", query,
			"error", param.ErrorMessage,
		)
	}
}

// handleSearch processes search requests
func handleSearch(searchManager SearchManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Validate request
		if strings.TrimSpace(req.Query) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
			return
		}

		// Build the search query
		query := map[string]any{
			"query": map[string]any{
				"match": map[string]any{
					"content": req.Query,
				},
			},
			"size": req.Size,
		}

		// Use the search manager to perform the search
		results, err := searchManager.Search(c.Request.Context(), req.Index, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the total count using a wrapped query
		total, err := searchManager.Count(c.Request.Context(), req.Index, map[string]any{
			"query": map[string]any{
				"match": map[string]any{
					"content": req.Query,
				},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Prepare and send the response
		response := SearchResponse{
			Results: results,
			Total:   int(total),
		}

		c.JSON(http.StatusOK, response)
	}
}

// StartHTTPServer starts the HTTP server for search requests
func StartHTTPServer(
	log logger.Interface,
	searchManager SearchManager,
	cfg config.Interface,
) (*http.Server, *middleware.SecurityMiddleware, error) {
	log.Info("StartHTTPServer function called")

	// Setup router
	router, security := SetupRouter(log, searchManager, cfg)

	// Get server config
	serverCfg := cfg.GetServerConfig()
	if serverCfg == nil {
		return nil, nil, errors.New("server configuration is required")
	}

	log.Info("Server configuration", "address", serverCfg.Address)

	// Create server
	server := &http.Server{
		Addr:              serverCfg.Address,
		Handler:           router,
		ReadTimeout:       serverCfg.ReadTimeout,
		WriteTimeout:      serverCfg.WriteTimeout,
		IdleTimeout:       serverCfg.IdleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return server, security, nil
}

// SetupLifecycle configures the lifecycle hooks for the API server
func SetupLifecycle(
	lc fx.Lifecycle,
	ctx context.Context,
	server *http.Server,
	searchManager SearchManager,
	security *middleware.SecurityMiddleware,
	log logger.Interface,
) {
	// Create a context for the cleanup goroutine using the provided context
	cleanupCtx, cancel := context.WithCancel(ctx)

	// Start the cleanup goroutine
	go security.Cleanup(cleanupCtx)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// No server start here - it's handled by httpd.go
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Create a timeout context for shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
			defer shutdownCancel()

			// Cancel the cleanup goroutine context
			cancel()

			// Wait for cleanup goroutine to finish with timeout
			cleanupDone := make(chan struct{})
			go func() {
				security.WaitCleanup()
				close(cleanupDone)
			}()

			select {
			case <-cleanupDone:
				// Cleanup completed successfully
			case <-shutdownCtx.Done():
				return nil // Return nil to indicate cleanup completed successfully
			}

			// Close the search manager
			if err := searchManager.Close(); err != nil {
				return fmt.Errorf("error closing search manager: %w", err)
			}

			return nil
		},
	})
}
