// Package api implements the HTTP API for the search service.
package api

import (
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"go.uber.org/fx"
)

const (
	// HealthCheckTimeout is the maximum time to wait for the server to become healthy
	HealthCheckTimeout = 5 * time.Second
	// HealthCheckInterval is the time between health check attempts
	HealthCheckInterval = 100 * time.Millisecond
	// ReadHeaderTimeout is the timeout for reading request headers
	ReadHeaderTimeout = 10 * time.Second
	// ShutdownTimeout is the timeout for graceful shutdown
	ShutdownTimeout = 5 * time.Second
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

// Module provides API dependencies
var Module = fx.Module("api",
	fx.Provide(
		// Provide the server and security middleware together to avoid circular dependencies
		func(
			log common.Logger,
			searchManager SearchManager,
			cfg common.Config,
		) (*http.Server, middleware.SecurityMiddlewareInterface) {
			// Create router and security middleware
			router, security := SetupRouter(log, searchManager, cfg)

			// Create server
			server := &http.Server{
				Addr:              cfg.GetServerConfig().Address,
				Handler:           router,
				ReadHeaderTimeout: ReadHeaderTimeout,
			}

			return server, security
		},
	),
	fx.Invoke(ConfigureLifecycle),
)
