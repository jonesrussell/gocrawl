// Package testutils provides test utilities for command modules.
package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/cmd/common/signal"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/common"
	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	sourcestest "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/jonesrussell/gocrawl/internal/storage"
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
	Sources  sources.Interface
	Storage  storage.Interface
	IndexMgr api.IndexManager
	Config   config.Interface
	Logger   logger.Logger
	Crawler  crawler.Interface

	// Command-specific dependencies
	Ctx            context.Context
	SourceName     string
	CommandDone    chan struct{}
	ArticleChannel chan *models.Article
	SignalHandler  *signal.SignalHandler
	Processors     []common.Processor
}

// NewCommandTestModule creates a new command test module with default values.
func NewCommandTestModule(t *testing.T) *CommandTestModule {
	// Set up mock logger
	mockLogger := &MockLogger{}
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
	mockCrawler := NewMockCrawler()

	return &CommandTestModule{
		Sources:        testSources,
		Storage:        NewMockStorage(),
		IndexMgr:       NewMockIndexManager(),
		Config:         mockConfig,
		Logger:         mockLogger,
		Crawler:        mockCrawler,
		Ctx:            t.Context(),
		SourceName:     "Test Source",
		CommandDone:    make(chan struct{}),
		ArticleChannel: make(chan *models.Article, crawler.ArticleChannelBufferSize),
		SignalHandler:  signal.NewSignalHandler(mockLogger),
		Processors:     []common.Processor{}, // Empty processors for testing
	}
}

// Module returns an fx.Option configured for command testing.
func (m *CommandTestModule) Module() fx.Option {
	return fx.Module("test",
		// Core dependencies
		fx.Provide(
			func() config.Interface { return m.Config },
			func() logger.Logger { return m.Logger },
			func() crawler.Interface { return m.Crawler },
			func() sources.Interface { return m.Sources },
			func() storage.Interface { return m.Storage },
			func() api.IndexManager { return m.IndexMgr },
		),

		// Command-specific dependencies
		fx.Provide(
			func() context.Context { return m.Ctx },
			func() string { return m.SourceName },
			func() chan struct{} { return m.CommandDone },
			func() chan *models.Article { return m.ArticleChannel },
			func() *signal.SignalHandler { return m.SignalHandler },
		),
		fx.Supply(m.Processors),
	)
}
