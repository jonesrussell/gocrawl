// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"net/http"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the HTTP server dependencies.
var Module = fx.Options(
	// Core modules
	config.Module,
	logger.Module,
	storage.Module,
	api.Module,

	// Providers
	fx.Provide(
		// Provide logger params
		func() logger.Params {
			return logger.Params{
				Config: &logger.Config{
					Level:       logger.InfoLevel,
					Development: false,
					Encoding:    "json",
				},
			}
		},
		// Provide config path
		func() string {
			return "config.yaml"
		},
	),

	// Invoke server startup
	fx.Invoke(func(lc fx.Lifecycle, srv *http.Server, log logger.Interface) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						log.Error("Server error", "error", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
	}),
)
