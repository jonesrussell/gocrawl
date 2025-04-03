// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the HTTP server command dependencies
var Module = fx.Module("httpd",
	config.TransportModule,
	config.Module,
	storage.Module,
	api.Module,
)
