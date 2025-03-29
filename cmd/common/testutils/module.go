// Package testutils provides test utilities for command modules.
package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	sourcestest "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	mockutils "github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
)

const (
	// Test configuration values
	testMaxDepth       = 3
	testParallelism    = 2
	testSourceMaxDepth = 2
)

// CommandTestModule provides a reusable test module for command tests.
type CommandTestModule struct {
	// Core dependencies
	Sources  sources.Interface `name:"testSourceManager"`
	Storage  types.Interface
	IndexMgr api.IndexManager
	Config   config.Interface
	Logger   common.Logger
	Crawler  crawler.Interface

	// Command-specific dependencies
	Ctx            context.Context       `name:"crawlContext"`
	SourceName     string                `name:"sourceName"`
	CommandDone    chan struct{}         `name:"commandDone"`
	ArticleChannel chan *models.Article  `name:"commandArticleChannel"`
	SignalHandler  *signal.SignalHandler `name:"signalHandler"`
	Processors     []collector.Processor `group:"processors"`
}

// NewCommandTestModule creates a new command test module with default values.
func NewCommandTestModule(t *testing.T) *CommandTestModule {
	// Set up mock logger
	mockLogger := &mockutils.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	// Set up test sources
	testConfigs := []sources.Config{
		{
			Name:      "Test Source",
			URL:       "http://test.example.com",
			RateLimit: time.Second,
			MaxDepth:  testSourceMaxDepth,
		},
	}
	testSources := sourcestest.NewTestSources(testConfigs)

	// Set up mock config
	mockConfig := configtestutils.NewMockConfig()

	// Set up mock crawler
	mockCrawler := mockutils.NewMockCrawler()

	return &CommandTestModule{
		Sources:        testSources,
		Storage:        mockutils.NewMockStorage(),
		IndexMgr:       mockutils.NewMockIndexManager(),
		Config:         mockConfig,
		Logger:         mockLogger,
		Crawler:        mockCrawler,
		Ctx:            t.Context(),
		SourceName:     "Test Source",
		CommandDone:    make(chan struct{}),
		ArticleChannel: make(chan *models.Article, crawler.ArticleChannelBufferSize),
		SignalHandler:  signal.NewSignalHandler(mockLogger),
		Processors:     []collector.Processor{}, // Empty processors for testing
	}
}

// Module returns an fx.Module configured for command testing.
func (m *CommandTestModule) Module() fx.Option {
	return fx.Module("test",
		// Core dependencies
		fx.Provide(
			// Provide mock config instead of using collector.Module
			fx.Annotate(
				func() config.Interface { return m.Config },
				fx.As(new(config.Interface)),
			),
			fx.Annotate(
				func() common.Logger { return m.Logger },
				fx.As(new(common.Logger)),
			),
			fx.Annotate(
				func() crawler.Interface { return m.Crawler },
				fx.As(new(crawler.Interface)),
			),
			collector.NewMetrics,
		),

		// Provide all required dependencies
		fx.Provide(
			// Command-specific dependencies
			fx.Annotate(
				func() context.Context { return m.Ctx },
				fx.ResultTags(`name:"crawlContext"`),
			),
			fx.Annotate(
				func() string { return m.SourceName },
				fx.ResultTags(`name:"sourceName"`),
			),
			fx.Annotate(
				func() chan struct{} { return m.CommandDone },
				fx.ResultTags(`name:"commandDone"`),
			),
			fx.Annotate(
				func() chan *models.Article { return m.ArticleChannel },
				fx.ResultTags(`name:"commandArticleChannel"`),
			),
			fx.Annotate(
				func() *signal.SignalHandler { return m.SignalHandler },
				fx.ResultTags(`name:"signalHandler"`),
			),
			// Test dependencies
			fx.Annotate(
				func() struct {
					fx.Out
					Sources    sources.Interface `name:"testSourceManager"`
					Storage    types.Interface
					IndexMgr   api.IndexManager      `name:"indexManager"`
					Processors []collector.Processor `group:"processors"`
				} {
					return struct {
						fx.Out
						Sources    sources.Interface `name:"testSourceManager"`
						Storage    types.Interface
						IndexMgr   api.IndexManager      `name:"indexManager"`
						Processors []collector.Processor `group:"processors"`
					}{
						Sources:    m.Sources,
						Storage:    m.Storage,
						IndexMgr:   m.IndexMgr,
						Processors: m.Processors,
					}
				},
			),
		),
	)
}
