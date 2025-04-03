// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// SearchManager defines the interface for search operations.
type SearchManager interface {
	// Search performs a search query.
	Search(ctx context.Context, index string, query map[string]interface{}) ([]interface{}, error)

	// Count returns the number of documents matching a query.
	Count(ctx context.Context, index string, query map[string]interface{}) (int64, error)

	// Aggregate performs an aggregation query.
	Aggregate(ctx context.Context, index string, aggs map[string]interface{}) (map[string]interface{}, error)

	// Close closes any resources held by the search manager.
	Close() error
}

// APIError represents an error response from the API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

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
) (*gin.Engine, middleware.SecurityMiddlewareInterface) {
	// Disable Gin's default logging
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(log))

	// Create security middleware
	security := middleware.NewSecurityMiddleware(cfg.GetServerConfig(), log)

	// Define public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Define protected routes
	protected := router.Group("")
	protected.Use(security.Middleware())
	{
		protected.POST("/search", handleSearch(searchManager))
	}

	return router, security
}

// loggingMiddleware creates a middleware that logs HTTP requests
func loggingMiddleware(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", statusCode,
			"latency", latency,
		)
	}
}

// handleSearch creates a handler for search requests
func handleSearch(searchManager SearchManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, APIError{
				Code:    http.StatusBadRequest,
				Message: "Invalid request payload",
				Err:     err,
			})
			return
		}

		// Validate request
		if strings.TrimSpace(req.Query) == "" {
			c.JSON(http.StatusBadRequest, APIError{
				Code:    http.StatusBadRequest,
				Message: "Query cannot be empty",
			})
			return
		}

		// Build the search query
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"content": req.Query,
				},
			},
			"size": req.Size,
		}

		// Use the search manager to perform the search
		results, err := searchManager.Search(c.Request.Context(), req.Index, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to perform search",
				Err:     err,
			})
			return
		}

		// Get the total count using a wrapped query
		total, err := searchManager.Count(c.Request.Context(), req.Index, map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"content": req.Query,
				},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to get total count",
				Err:     err,
			})
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
) (*http.Server, middleware.SecurityMiddlewareInterface, error) {
	router, security := SetupRouter(log, searchManager, cfg)

	srv := &http.Server{
		Addr:              cfg.GetServerConfig().Address,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       cfg.GetServerConfig().ReadTimeout,
		WriteTimeout:      cfg.GetServerConfig().WriteTimeout,
		IdleTimeout:       cfg.GetServerConfig().IdleTimeout,
	}

	return srv, security, nil
}
