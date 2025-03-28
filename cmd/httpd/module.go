// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"go.uber.org/fx"
)

// searchManagerWrapper wraps types.Interface to implement api.SearchManager
type searchManagerWrapper struct {
	types.Interface
}

func (w *searchManagerWrapper) Close() error {
	return nil
}

// Module provides the HTTP server command dependencies
var Module = fx.Module("httpd",
	api.Module,
	fx.Provide(
		// Provide a SearchManager implementation
		func(storage types.Interface) api.SearchManager {
			return &searchManagerWrapper{storage}
		},
	),
)
