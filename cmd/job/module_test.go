// Package job_test implements tests for the job scheduler command.
package job_test

import (
	"context"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/cmd/job"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockCrawler is a mock implementation of crawler.Interface for testing
type mockCrawler struct{}

func (m *mockCrawler) Start(_ context.Context, _ string) error { return nil }
func (m *mockCrawler) Stop(_ context.Context) error            { return nil }
func (m *mockCrawler) Wait()                                   {}
func (m *mockCrawler) SetCollector(_ *colly.Collector)         {}
func (m *mockCrawler) SetMaxDepth(_ int)                       {}
func (m *mockCrawler) SetRateLimit(_ string) error             { return nil }
func (m *mockCrawler) GetIndexManager() api.IndexManager       { return nil }
func (m *mockCrawler) Subscribe(_ events.Handler)              {}

// mockLogger is a mock implementation of logger.Interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(_ string, _ ...any)       {}
func (m *mockLogger) Info(_ string, _ ...any)        {}
func (m *mockLogger) Warn(_ string, _ ...any)        {}
func (m *mockLogger) Error(_ string, _ ...any)       {}
func (m *mockLogger) Fatal(_ string, _ ...any)       {}
func (m *mockLogger) Panic(_ string, _ ...any)       {}
func (m *mockLogger) With(_ ...any) logger.Interface { return m }
func (m *mockLogger) Errorf(_ string, _ ...any)      {}
func (m *mockLogger) Printf(_ string, _ ...any)      {}
func (m *mockLogger) Sync() error                    { return nil }

// mockConfig is a mock implementation of config.Interface for testing
type mockConfig struct{}

func (m *mockConfig) GetCrawlerConfig() *config.CrawlerConfig { return &config.CrawlerConfig{} }
func (m *mockConfig) GetElasticsearchConfig() *config.ElasticsearchConfig {
	return &config.ElasticsearchConfig{}
}
func (m *mockConfig) GetLogConfig() *config.LogConfig       { return &config.LogConfig{} }
func (m *mockConfig) GetAppConfig() *config.AppConfig       { return &config.AppConfig{} }
func (m *mockConfig) GetServerConfig() *config.ServerConfig { return &config.ServerConfig{} }
func (m *mockConfig) GetSources() []config.Source           { return []config.Source{} }

// TestModuleProvides tests that the job module provides all necessary dependencies
func TestModuleProvides(t *testing.T) {
	var command *cobra.Command

	app := fxtest.New(t,
		job.Module,
		fx.Provide(
			// Mock dependencies
			func() logger.Interface { return &mockLogger{} },
			func() config.Interface { return &mockConfig{} },
			func() *sources.Sources { return &sources.Sources{} },
			func() crawler.Interface { return &mockCrawler{} },
		),
		fx.Populate(&command),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, command)
	require.Equal(t, "job", command.Use)
}

// TestModuleConfiguration tests the module's configuration behavior
func TestModuleConfiguration(t *testing.T) {
	var command *cobra.Command

	app := fxtest.New(t,
		job.Module,
		fx.Provide(
			// Mock dependencies
			func() logger.Interface { return &mockLogger{} },
			func() config.Interface { return &mockConfig{} },
			func() *sources.Sources { return &sources.Sources{} },
			func() crawler.Interface { return &mockCrawler{} },
		),
		fx.Populate(&command),
	)
	defer app.RequireStart().RequireStop()

	require.NotNil(t, command)
	require.Equal(t, "job", command.Use)
	require.Equal(t, "Schedule and run crawl jobs", command.Short)
	require.Contains(t, command.Long, "Schedule and run crawl jobs based on the times specified in sources.yml")
}
