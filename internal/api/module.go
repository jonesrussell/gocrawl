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

// APIParams contains the dependencies for API constructors
type APIParams struct {
	fx.In

	Logger  logger.Interface
	Storage types.Interface
}

// SearchResult contains the search manager result
type SearchResult struct {
	fx.Out

	Search Search
}

// IndexResult contains the index manager result
type IndexResult struct {
	fx.Out

	IndexManager IndexManager
}

// DocumentResult contains the document manager result
type DocumentResult struct {
	fx.Out

	DocumentManager DocumentManager
}

// searchManager implements the Search interface
type searchManager struct {
	storage types.Interface
	logger  logger.Interface
}

func (m *searchManager) Search(ctx context.Context, query string) ([]any, error) {
	return m.storage.Search(ctx, "articles", map[string]any{"query": query})
}

// NewSearchManager creates a new search manager
func NewSearchManager(p APIParams) SearchResult {
	return SearchResult{
		Search: &searchManager{
			storage: p.Storage,
			logger:  p.Logger,
		},
	}
}

// NewIndexManager creates a new index manager
func NewIndexManager(p APIParams) IndexResult {
	return IndexResult{
		IndexManager: p.Storage.GetIndexManager(),
	}
}

// documentManager implements the DocumentManager interface
type documentManager struct {
	storage types.Interface
	logger  logger.Interface
}

func (m *documentManager) Index(ctx context.Context, index string, id string, doc any) error {
	return m.storage.IndexDocument(ctx, index, id, doc)
}

func (m *documentManager) Update(ctx context.Context, index string, id string, doc any) error {
	return m.storage.IndexDocument(ctx, index, id, doc)
}

func (m *documentManager) Delete(ctx context.Context, index string, id string) error {
	return m.storage.DeleteDocument(ctx, index, id)
}

func (m *documentManager) Get(ctx context.Context, index string, id string) (any, error) {
	var doc any
	err := m.storage.GetDocument(ctx, index, id, &doc)
	return doc, err
}

// NewDocumentManager creates a new document manager
func NewDocumentManager(p APIParams) DocumentResult {
	return DocumentResult{
		DocumentManager: &documentManager{
			storage: p.Storage,
			logger:  p.Logger,
		},
	}
}

// Module provides the API module for dependency injection.
var Module = fx.Module("api",
	fx.Provide(
		NewSearchManager,
		NewIndexManager,
		NewDocumentManager,
	),
)

// NewAPI creates a new API instance.
func NewAPI(
	ctx context.Context,
	cfg config.Interface,
	log logger.Interface,
	storage types.Interface,
	indexManager types.IndexManager,
) *Server {
	return &Server{
		Context:      ctx,
		Config:       cfg,
		Logger:       log,
		Storage:      storage,
		IndexManager: indexManager,
	}
}
