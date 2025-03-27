package crawl_test

import (
	"context"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/cmd/crawl"
	"github.com/jonesrussell/gocrawl/internal/api"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/crawler/events"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/models"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/storage/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig implements config.Interface for testing
type mockConfig struct {
	config.Interface
}

// mockStorageImpl implements types.Interface for testing
type mockStorageImpl struct {
	types.Interface
}

func (m *mockStorageImpl) TestConnection(ctx context.Context) error { return nil }
func (m *mockStorageImpl) Close() error                             { return nil }

// mockCrawler implements crawler.Interface for testing
type mockCrawler struct {
	crawler.Interface
	indexManager api.IndexManager
}

func (m *mockCrawler) Start(ctx context.Context, url string) error { return nil }
func (m *mockCrawler) Stop(ctx context.Context) error              { return nil }
func (m *mockCrawler) Wait()                                       {}
func (m *mockCrawler) Subscribe(handler events.Handler)            {}
func (m *mockCrawler) SetRateLimit(duration string) error          { return nil }
func (m *mockCrawler) SetMaxDepth(depth int)                       {}
func (m *mockCrawler) SetCollector(collector *colly.Collector)     {}
func (m *mockCrawler) GetIndexManager() api.IndexManager           { return m.indexManager }

// mockSources implements sources.Interface for testing
type mockSources struct {
	sources.Interface
}

// mockArticleService implements article.Interface for testing
type mockArticleService struct {
	article.Interface
}

// mockContentService implements content.Interface for testing
type mockContentService struct {
	content.Interface
}

// mockIndexManagerImpl implements api.IndexManager for testing
type mockIndexManagerImpl struct {
	api.IndexManager
}

func (m *mockIndexManagerImpl) EnsureIndex(ctx context.Context, name string, mapping any) error {
	return nil
}
func (m *mockIndexManagerImpl) DeleteIndex(ctx context.Context, name string) error { return nil }
func (m *mockIndexManagerImpl) IndexExists(ctx context.Context, name string) (bool, error) {
	return true, nil
}
func (m *mockIndexManagerImpl) UpdateMapping(ctx context.Context, name string, mapping any) error {
	return nil
}

func TestAppDependencies(t *testing.T) {
	t.Parallel()

	indexManager := &mockIndexManagerImpl{}
	articleChan := make(chan *models.Article)
	doneChan := make(chan struct{})

	app := fxtest.New(t,
		fx.NopLogger,
		crawl.Module,
		fx.Provide(
			logger.NewNoOp,
			func() config.Interface { return &mockConfig{} },
			func() types.Interface { return &mockStorageImpl{} },
			func() crawler.Interface { return &mockCrawler{indexManager: indexManager} },
			func() sources.Interface { return &mockSources{} },
			func() api.IndexManager { return indexManager },
			func() context.Context { return t.Context() },
			func() article.Interface { return &mockArticleService{} },
			func() content.Interface { return &mockContentService{} },
			fx.Annotate(
				func() chan *models.Article { return articleChan },
				fx.ResultTags(`name:"articleChannel"`),
			),
			fx.Annotate(
				func() chan struct{} { return doneChan },
				fx.ResultTags(`name:"crawlDone"`),
			),
			fx.Annotate(
				func() string { return "test-source" },
				fx.ResultTags(`name:"sourceName"`),
			),
		),
		fx.Invoke(func(p crawl.Dependencies) {
			assert.NotNil(t, p.Logger)
			assert.NotNil(t, p.Config)
			assert.NotNil(t, p.Storage)
			assert.NotNil(t, p.Crawler)
			assert.NotNil(t, p.Sources)
			assert.NotNil(t, p.Context)
		}),
	)

	app.RequireStart()
	app.RequireStop()
}
