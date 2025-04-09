// Package test provides test utilities for the indices command.
package test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/app"
	"github.com/jonesrussell/gocrawl/internal/config/log"
	"github.com/jonesrussell/gocrawl/internal/config/priority"
	"github.com/jonesrussell/gocrawl/internal/config/server"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"go.uber.org/fx"
)

// TestModule provides a test module with mock dependencies
func TestModule(t *testing.T) fx.Option {
	return fx.Module("test",
		fx.Provide(
			func() context.Context { return t.Context() },
			func() config.Interface {
				mockCfg := &configtestutils.MockConfig{}
				mockCfg.On("GetAppConfig").Return(&app.Config{
					Environment: "test",
					Name:        "gocrawl",
					Version:     "1.0.0",
					Debug:       true,
				})
				mockCfg.On("GetLogConfig").Return(&log.Config{
					Level: "debug",
				})
				mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
					IndexName: "test-index",
				})
				mockCfg.On("GetServerConfig").Return(&server.Config{
					Address: ":8080",
				})
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				mockCfg.On("GetPriorityConfig").Return(&priority.Config{
					Default: 1,
					Rules:   []priority.Rule{},
				})
				return mockCfg
			},
			logger.NewNoOp,
		),
	)
}
