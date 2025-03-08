// Package indexing provides functionality for indexing and searching documents.
package indexing

import (
	"context"

	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/indexing/client"
	"github.com/jonesrussell/gocrawl/internal/storage/elasticsearch"
	"go.uber.org/fx"
)

// Params defines the dependencies required by the indexing package.
type Params struct {
	fx.In

	Config *config.Config
	Logger common.Logger
}

// Result contains the services provided by the indexing package.
type Result struct {
	fx.Out

	IndexManager    api.IndexManager    `group:"index"`
	DocumentManager api.DocumentManager `group:"document"`
	SearchManager   api.SearchManager   `group:"search"`
}

// Module provides the indexing services.
var Module = fx.Options(
	fx.Provide(
		provideClient,
		provideIndexManager,
		provideDocumentManager,
		provideSearchManager,
	),
	fx.Invoke(RegisterHooks),
)

func provideClient(p Params) (*client.Client, error) {
	return client.New(p.Config, p.Logger)
}

func provideIndexManager(client *client.Client, logger common.Logger) (api.IndexManager, error) {
	return elasticsearch.NewManager(client, logger)
}

func provideDocumentManager(client *client.Client, logger common.Logger) (api.DocumentManager, error) {
	return elasticsearch.NewDocumentManager(client, logger)
}

func provideSearchManager(client *client.Client, logger common.Logger) (api.SearchManager, error) {
	return elasticsearch.NewSearchManager(client, logger)
}

// RegisterHooks registers lifecycle hooks for the indexing module.
func RegisterHooks(lc fx.Lifecycle, _ api.IndexManager, logger common.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Starting indexing module")
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Stopping indexing module")
			return nil
		},
	})
}
