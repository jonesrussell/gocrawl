// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/storage"
	"go.uber.org/fx"
)

// Module provides the HTTP server dependencies.
var Module = fx.Options(
	// Core modules
	logger.Module,
	storage.Module,
	api.Module,

	// Providers
	fx.Provide(
		// Provide the router and security middleware
		func(
			cfg config.Interface,
			log logger.Interface,
			searchManager api.SearchManager,
		) (*gin.Engine, middleware.SecurityMiddlewareInterface, error) {
			router, security := api.SetupRouter(log, searchManager, cfg)
			return router, security, nil
		},
		// Provide the server
		func(
			cfg config.Interface,
			log logger.Interface,
			router *gin.Engine,
			security middleware.SecurityMiddlewareInterface,
		) (*http.Server, error) {
			srv := &http.Server{
				Addr:              cfg.GetServerConfig().Address,
				Handler:           router,
				ReadTimeout:       cfg.GetServerConfig().ReadTimeout,
				WriteTimeout:      cfg.GetServerConfig().WriteTimeout,
				IdleTimeout:       cfg.GetServerConfig().IdleTimeout,
				ReadHeaderTimeout: api.ReadHeaderTimeout,
			}
			return srv, nil
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
