// Package api implements the HTTP API for the search service.
package api

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	storagetypes "github.com/jonesrussell/gocrawl/internal/storage/types"
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
	Storage      storagetypes.Interface
	IndexManager storagetypes.IndexManager
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
