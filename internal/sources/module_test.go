package sources

import (
	"context"
	"os"
	"testing"

	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func createTestFile(t *testing.T) string {
	content := `sources:
  - name: test
    url: http://example.com
    index: test_index
    rate_limit: 1s
    max_depth: 2
    time:
      - "09:00"
      - "15:00"`

	tmpfile, err := os.CreateTemp("", "sources*.yml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name()
}

func createTestModule(configPath string) fx.Option {
	return fx.Module("sources",
		fx.Provide(
			// Provide the sources configuration with crawler
			func(logger logger.Interface, c *crawler.Crawler) (*Sources, error) {
				sources, err := Load(configPath)
				if err != nil {
					return nil, err
				}
				sources.Logger = logger
				sources.Crawler = c
				return sources, nil
			},
			// Provide individual source configs
			fx.Annotate(
				func(s *Sources) []Config {
					return s.Sources
				},
				fx.ResultTags(`group:"sources"`),
			),
		),
	)
}

func TestModule(t *testing.T) {
	configPath := createTestFile(t)

	// Create test app with required dependencies
	app := fxtest.New(t,
		fx.Provide(
			func() logger.Interface {
				return new(MockLogger)
			},
			func() *crawler.Crawler {
				return &crawler.Crawler{}
			},
		),
		createTestModule(configPath),
		fx.Invoke(func(s *Sources) {
			// Verify sources are loaded correctly
			assert.NotNil(t, s)
			assert.NotNil(t, s.Logger)
			assert.NotNil(t, s.Crawler)
			assert.Len(t, s.Sources, 1)

			// Verify source config
			source := s.Sources[0]
			assert.Equal(t, "test", source.Name)
			assert.Equal(t, "http://example.com", source.URL)
			assert.Equal(t, "test_index", source.Index)
			assert.Equal(t, "1s", source.RateLimit)
			assert.Equal(t, 2, source.MaxDepth)
			assert.Equal(t, []string{"09:00", "15:00"}, source.Time)
		}),
	)

	require.NoError(t, app.Start(context.Background()))
	defer app.Stop(context.Background())
}
