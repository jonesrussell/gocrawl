package api

import (
	"context"
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

// setupRouter creates and configures the Gin router with all routes
func setupRouter(log logger.Interface, searchManager SearchManager, cfg config.Interface) *gin.Engine {
	router := gin.New()

	// Create security middleware
	security := middleware.NewSecurityMiddleware(cfg.GetServerConfig(), log)
	router.Use(security.Middleware())

	// Define routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/search", handleSearch(searchManager))

	return router
}

// handleSearch processes search requests
func handleSearch(searchManager SearchManager) gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}

// getServerAddress determines the server address with priority:
// 1. GOCRAWL_PORT environment variable
// 2. Server config from config file
// 3. Default port
func getServerAddress(cfg config.Interface) string {
	var port string
	if envPort := os.Getenv("GOCRAWL_PORT"); envPort != "" {
		port = envPort
	} else if serverCfg := cfg.GetServerConfig(); serverCfg != nil && serverCfg.Address != "" {
		port = strings.TrimPrefix(serverCfg.Address, ":")
	} else {
		port = defaultPort
	}

	// Ensure port has colon prefix
	return ":" + strings.TrimPrefix(port, ":")
}

// configureTLS sets up TLS configuration if enabled
func configureTLS(server *http.Server, cfg config.Interface, log logger.Interface) error {
	if !cfg.GetServerConfig().Security.TLS.Enabled {
		log.Info("TLS is disabled")
		return nil
	}

	log.Info("TLS is enabled, loading certificates",
		"certificate", cfg.GetServerConfig().Security.TLS.Certificate,
		"key", cfg.GetServerConfig().Security.TLS.Key)

	// Validate TLS configuration at startup
	cert, err := tls.LoadX509KeyPair(
		cfg.GetServerConfig().Security.TLS.Certificate,
		cfg.GetServerConfig().Security.TLS.Key,
	)
	if err != nil {
		log.Error("Failed to load TLS certificate", "error", err)
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	log.Info("TLS certificate loaded successfully")

	server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			log.Debug("Getting certificate for client", "server_name", hello.ServerName)
			return &cert, nil
		},
	}
	log.Info("TLS configuration completed")
	return nil
}

// StartHTTPServer starts the HTTP server for search requests
func StartHTTPServer(log logger.Interface, searchManager SearchManager, cfg config.Interface) (*http.Server, error) {
	log.Info("StartHTTPServer function called")

	// Setup router
	router := setupRouter(log, searchManager, cfg)

	// Get server address
	address := getServerAddress(cfg)
	log.Info("Server configuration", "address", address)

	// Create server
	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadTimeout:       cfg.GetServerConfig().ReadTimeout,
		WriteTimeout:      cfg.GetServerConfig().WriteTimeout,
		IdleTimeout:       cfg.GetServerConfig().IdleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Configure TLS if enabled
	if err := configureTLS(server, cfg, log); err != nil {
		return nil, err
	}

	return server, nil
}

// Module provides API dependencies
var Module = fx.Module("api",
	fx.Provide(
		StartHTTPServer,
	),
	fx.Invoke(func(lc fx.Lifecycle, server *http.Server, cfg config.Interface) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Start server in a goroutine
				go func() {
					var err error
					if cfg.GetServerConfig().Security.TLS.Enabled {
						err = server.ListenAndServeTLS("", "") // Certificates are already loaded
					} else {
						err = server.ListenAndServe()
					}
					if err != nil && err != http.ErrServerClosed {
						panic(err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.Shutdown(ctx)
			},
		})
	}),
)
