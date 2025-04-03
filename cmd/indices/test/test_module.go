// Package test provides test utilities for the indices command.
package test

import (
	"context"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
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
				mockCfg.On("GetAppConfig").Return(&config.AppConfig{
					Environment: "test",
					Name:        "gocrawl",
					Version:     "1.0.0",
					Debug:       true,
				})
				mockCfg.On("GetLogConfig").Return(&config.LogConfig{
					Level: "debug",
					Debug: true,
				})
				mockCfg.On("GetElasticsearchConfig").Return(&config.ElasticsearchConfig{
					Addresses: []string{"http://localhost:9200"},
					IndexName: "test-index",
				})
				mockCfg.On("GetServerConfig").Return(&config.ServerConfig{
					Address: ":8080",
				})
				mockCfg.On("GetSources").Return([]config.Source{}, nil)
				mockCfg.On("GetCommand").Return("test")
				mockCfg.On("GetPriorityConfig").Return(&config.PriorityConfig{
					Default: 1,
					Rules:   []config.PriorityRule{},
				})
				return mockCfg
			},
			logger.NewNoOp,
		),
	)
}
