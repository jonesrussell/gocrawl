// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"go.uber.org/fx"
)

// Module provides the HTTP server dependencies.
// Note: Command registration is handled by Command() function in httpd.go, not through FX Group annotation.
// The server is started manually in the command's RunE function, not via FX lifecycle hooks.
var Module = fx.Module("httpd",
	// Core modules
	common.Module,
	api.Module,
)
