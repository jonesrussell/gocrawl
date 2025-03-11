// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"github.com/jonesrussell/gocrawl/internal/api"
	"go.uber.org/fx"
)

// Module provides the HTTP server command dependencies
var Module = fx.Module("httpd",
	api.Module,
)
