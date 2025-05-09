// Package api implements the HTTP API for the search service.
package api

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
)

// Server represents the API server.
type Server struct {
	Context      context.Context
	Config       config.Interface
	Logger       logger.Interface
	Storage      types.Interface
	IndexManager types.IndexManager
}

// Params holds the parameters for creating a new API server.
type Params struct {
	Context      context.Context
	Config       config.Interface
	Logger       logger.Interface
	Storage      types.Interface
	IndexManager types.IndexManager
}

// NewServer creates a new API server instance.
func NewServer(p Params) *Server {
	return &Server{
		Context:      p.Context,
		Config:       p.Config,
		Logger:       p.Logger,
		Storage:      p.Storage,
		IndexManager: p.IndexManager,
	}
}

// Start starts the API server.
func (s *Server) Start(ctx context.Context) error {
	s.Logger.Info("Starting API server")
	return nil
}

// Stop stops the API server.
func (s *Server) Stop(ctx context.Context) error {
	s.Logger.Info("Stopping API server")
	return nil
}
