// Package api provides the API layer for the application.
package api

import (
	"context"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
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

// searchManager implements the Search interface
type searchManager struct {
	storage types.Interface
	logger  logger.Interface
}

func (m *searchManager) Search(ctx context.Context, query string) ([]any, error) {
	return m.storage.Search(ctx, "articles", map[string]any{"query": query})
}

// documentManager implements the DocumentManager interface
type documentManager struct {
	storage types.Interface
	logger  logger.Interface
}

// Index indexes a document
func (m *documentManager) Index(ctx context.Context, index, id string, doc any) error {
	return m.storage.IndexDocument(ctx, index, id, doc)
}

// Update updates a document
func (m *documentManager) Update(ctx context.Context, index, id string, doc any) error {
	return m.storage.IndexDocument(ctx, index, id, doc)
}

// Delete deletes a document
func (m *documentManager) Delete(ctx context.Context, index, id string) error {
	return m.storage.DeleteDocument(ctx, index, id)
}

// Get gets a document
func (m *documentManager) Get(ctx context.Context, index, id string) (any, error) {
	var doc any
	err := m.storage.GetDocument(ctx, index, id, &doc)
	return doc, err
}

// ProvideAPI creates all API components
func ProvideAPI(p struct {
	fx.In

	Config       config.Interface
	Logger       logger.Interface
	Storage      types.Interface
	IndexManager types.IndexManager
}) (struct {
	fx.Out

	Search          Search
	IndexManager    IndexManager
	DocumentManager DocumentManager
	Server          *Server
}, error) {
	// Create search manager
	search := &searchManager{
		storage: p.Storage,
		logger:  p.Logger,
	}

	// Create document manager
	docManager := &documentManager{
		storage: p.Storage,
		logger:  p.Logger,
	}

	// Create server
	server := &Server{
		Context:      context.Background(),
		Config:       p.Config,
		Logger:       p.Logger,
		Storage:      p.Storage,
		IndexManager: p.IndexManager,
	}

	return struct {
		fx.Out

		Search          Search
		IndexManager    IndexManager
		DocumentManager DocumentManager
		Server          *Server
	}{
		Search:          search,
		IndexManager:    p.IndexManager,
		DocumentManager: docManager,
		Server:          server,
	}, nil
}

// Module provides the API module for dependency injection.
var Module = fx.Module("api",
	fx.Provide(
		ProvideAPI,
	),
)
