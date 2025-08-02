// Package httpd implements the HTTP server command for the search API.
package httpd

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonesrussell/gocrawl/cmd/common"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/api/middleware"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Module provides the HTTP server dependencies.
var Module = fx.Module("httpd",
	// Core modules
	common.Module,
	api.Module,

	// Providers
	fx.Provide(
		// Provide the router and security middleware
		func(
			cfg config.Interface,
			log logger.Interface,
		) (*gin.Engine, middleware.SecurityMiddlewareInterface, error) {
			router, security := api.SetupRouter(log, nil, cfg)
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
		// Provide the command registrar
		fx.Annotated{
			Group: "commands",
			Target: func(
				cfg config.Interface,
				log logger.Interface,
				srv *http.Server,
			) common.CommandRegistrar {
				return func(parent *cobra.Command) {
					cmd := &cobra.Command{
						Use:   "serve",
						Short: "Start the HTTP server",
						Long:  `Start the HTTP server for the search API`,
						RunE: func(cmd *cobra.Command, args []string) error {
							// Start the server
							if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
								return err
							}
							return nil
						},
					}

					parent.AddCommand(cmd)
				}
			},
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
