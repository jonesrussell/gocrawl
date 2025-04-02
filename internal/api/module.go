// Package api implements the HTTP API for the search service.
package api

import (
	"net/http"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/common/types"
	"github.com/jonesrussell/gocrawl/internal/logger"
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
			log types.Logger,
			searchManager SearchManager,
			cfg common.Config,
		) (*http.Server, middleware.SecurityMiddlewareInterface) {
			// Use StartHTTPServer to create the server and security middleware
			server, security, err := StartHTTPServer(log, searchManager, cfg)
			if err != nil {
				panic(err)
			}
			return server, security
		},
	),
	fx.Invoke(ConfigureLifecycle),
	logger.Module,
)
