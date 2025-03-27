// Package api implements the HTTP API for the search service.
package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
)

// Constants
const (
	readHeaderTimeout = 10 * time.Second // Timeout for reading headers
	shutdownTimeout   = 5 * time.Second  // Timeout for graceful shutdown
)

// API implements the HTTP API
type API struct {
	router *gin.Engine
	config *config.Config
	log    common.Logger
}

// NewAPI creates a new API instance
func NewAPI(router *gin.Engine, config *config.Config, log common.Logger) *API {
	return &API{
		router: router,
		config: config,
		log:    log,
	}
}

// SetupRouter creates and configures the Gin router with all routes
func SetupRouter(
	log common.Logger,
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
	router.Use(security.Middleware())

	// Define routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/search", handleSearch(searchManager))

	return router, security
}

// loggingMiddleware creates a middleware that logs HTTP requests
func loggingMiddleware(log common.Logger) gin.HandlerFunc {
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
	log common.Logger,
	searchManager SearchManager,
	cfg config.Interface,
) (*http.Server, middleware.SecurityMiddlewareInterface, error) {
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
