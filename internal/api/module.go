package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
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
	defaultPort       = "8080"           // Default port if not specified in config or env
)

// SearchRequest represents the structure of the search request
type SearchRequest struct {
	Query string `json:"query"`
	Index string `json:"index"`
	Size  int    `json:"size"`
}

// SearchResponse represents the structure of the search response
type SearchResponse struct {
	Results []any `json:"results"`
	Total   int   `json:"total"`
}

// StartHTTPServer starts the HTTP server for search requests
func StartHTTPServer(log logger.Interface, searchManager SearchManager, cfg config.Interface) (*http.Server, error) {
	log.Info("StartHTTPServer function called")

	// Create Gin router
	router := gin.New()

	// Create security middleware
	security := middleware.NewSecurityMiddleware(cfg.GetServerConfig(), log)
	router.Use(security.Middleware())

	// Define routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/search", func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
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
	})

	// Determine server address with priority:
	// 1. GOCRAWL_PORT environment variable
	// 2. Server config from config file
	// 3. Default port
	var port string
	if envPort := os.Getenv("GOCRAWL_PORT"); envPort != "" {
		port = envPort
	} else if serverCfg := cfg.GetServerConfig(); serverCfg != nil && serverCfg.Address != "" {
		port = strings.TrimPrefix(serverCfg.Address, ":")
	} else {
		port = defaultPort
	}

	// Ensure port has colon prefix
	address := ":" + strings.TrimPrefix(port, ":")

	log.Info("Server configuration", "address", address)

	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadTimeout:       cfg.GetServerConfig().ReadTimeout,
		WriteTimeout:      cfg.GetServerConfig().WriteTimeout,
		IdleTimeout:       cfg.GetServerConfig().IdleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Configure TLS if enabled
	if cfg.GetServerConfig().Security.TLS.Enabled {
		// Validate TLS configuration at startup
		cert, err := tls.LoadX509KeyPair(
			cfg.GetServerConfig().Security.TLS.Certificate,
			cfg.GetServerConfig().Security.TLS.Key,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}

		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &cert, nil
			},
		}
	}

	return server, nil
}

// Module provides API dependencies
var Module = fx.Module("api",
	fx.Provide(
		StartHTTPServer,
	),
)
