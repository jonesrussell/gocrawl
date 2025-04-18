// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/interfaces"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
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

// Module provides the API dependencies.
var Module = fx.Module("api",
	fx.Provide(
		// Provide the router and security middleware
		func(
			cfg config.Interface,
			log logger.Interface,
			searchManager SearchManager,
		) (*gin.Engine, middleware.SecurityMiddlewareInterface, error) {
			router, security := SetupRouter(log, searchManager, cfg)
			return router, security, nil
		},
		NewLifecycle,
		NewServer,
	),
)

// Params holds the dependencies required for the API.
type Params struct {
	fx.In
	Context      context.Context `name:"apiContext"`
	Config       config.Interface
	Logger       logger.Interface
	Storage      types.Interface
	IndexManager interfaces.IndexManager
}

// NewAPI creates a new API instance.
func NewAPI(p Params) *Server {
	return &Server{
		Context:      p.Context,
		Config:       p.Config,
		Logger:       p.Logger,
		Storage:      p.Storage,
		IndexManager: p.IndexManager,
	}
}
