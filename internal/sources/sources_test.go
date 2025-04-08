// Package sources_test provides tests for the sources package.
package sources_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/api/testutils"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	sourcetestutils "github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// testSetup holds common test data and utilities
type testSetup struct {
	t           *testing.T
	tmpDir      string
	sourcesFile string
	config      *config.Config
}

// newTestSetup creates a new test setup with temporary files
func newTestSetup(t *testing.T) *testSetup {
	tmpDir := t.TempDir()
	sourcesFile := filepath.Join(tmpDir, "sources.yml")

	// Create test sources file
	testSources := `
sources:
  - name: test-source
    url: https://example.com
    rate_limit: 1s
    max_depth: 2
    selectors:
      article:
        title: h1
        body: article
`
	err := os.WriteFile(sourcesFile, []byte(testSources), 0644)
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		Crawler: &config.CrawlerConfig{
			SourceFile: sourcesFile,
		},
	}

	return &testSetup{
		t:           t,
		tmpDir:      tmpDir,
		sourcesFile: sourcesFile,
		config:      cfg,
	}
}

// TestLoadFromFile tests loading sources from a file
func TestLoadFromFile(t *testing.T) {
	t.Parallel()
	setup := newTestSetup(t)

	t.Run("valid sources file", func(t *testing.T) {
		t.Parallel()
		loaderConfigs, err := loader.LoadFromFile(setup.sourcesFile)
		require.NoError(t, err)
		require.Len(t, loaderConfigs, 1)

		sourceConfigs := make([]sourceutils.SourceConfig, len(loaderConfigs))
		for i, cfg := range loaderConfigs {
			rateLimit, parseErr := time.ParseDuration(cfg.RateLimit)
			require.NoError(t, parseErr)

			sourceConfigs[i] = sourceutils.SourceConfig{
				Name:      cfg.Name,
				URL:       cfg.URL,
				RateLimit: rateLimit,
				MaxDepth:  cfg.MaxDepth,
				Selectors: sourceutils.SelectorConfig{
					Article: sourceutils.ArticleSelectors{
						Title: cfg.Selectors.Article.Title,
						Body:  cfg.Selectors.Article.Body,
					},
				},
			}
		}

		s := sourcetestutils.NewTestSources(sourceConfigs)
		require.NotNil(t, s)

		sources, err := s.ListSources(t.Context())
		require.NoError(t, err)
		require.Len(t, sources, 1)
		assert.Equal(t, "test-source", sources[0].Name)
		assert.Equal(t, "https://example.com", sources[0].URL)
		assert.Equal(t, time.Second, sources[0].RateLimit)
		assert.Equal(t, 2, sources[0].MaxDepth)
		assert.Equal(t, "h1", sources[0].Selectors.Article.Title)
		assert.Equal(t, "article", sources[0].Selectors.Article.Body)
	})

	t.Run("invalid sources file", func(t *testing.T) {
		invalidFile := filepath.Join(setup.tmpDir, "invalid.yml")
		err := os.WriteFile(invalidFile, []byte("invalid: yaml: content"), 0644)
		require.NoError(t, err)

		_, err = loader.LoadFromFile(invalidFile)
		require.Error(t, err)
	})
}

// TestGetSource tests the GetSource method
func TestGetSource(t *testing.T) {
	t.Parallel()

	t.Run("existing source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{
			{
				Name:      "test-source",
				URL:       "https://test.com",
				RateLimit: 1,
				MaxDepth:  2,
			},
		})
		source := s.FindByName("test-source")
		require.NotNil(t, source)
		assert.Equal(t, "test-source", source.Name)
	})

	t.Run("non-existent source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{})
		source := s.FindByName("non-existent")
		require.Nil(t, source)
	})
}

// TestValidateSource tests source validation
func TestValidateSource(t *testing.T) {
	t.Parallel()

	t.Run("valid source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources(nil)
		source := &sourceutils.SourceConfig{
			Name:      "test-source",
			URL:       "https://example.com",
			RateLimit: time.Second,
			MaxDepth:  2,
			Selectors: sourceutils.SelectorConfig{
				Article: sourceutils.ArticleSelectors{
					Title: "h1",
					Body:  "article",
				},
			},
		}
		err := s.ValidateSource(source)
		require.NoError(t, err)
	})
	t.Run("invalid source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources(nil)
		source := &sourceutils.SourceConfig{
			Name:      "",
			URL:       "",
			RateLimit: 0,
			MaxDepth:  0,
		}
		err := s.ValidateSource(source)
		require.Error(t, err)
	})
}

// TestSourceOperations tests source CRUD operations
func TestSourceOperations(t *testing.T) {
	t.Parallel()

	t.Run("add source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{})
		err := s.AddSource(t.Context(), &sourceutils.SourceConfig{
			Name:      "test-source",
			URL:       "https://example.com",
			RateLimit: time.Second,
			MaxDepth:  2,
			Selectors: sourceutils.SelectorConfig{
				Article: sourceutils.ArticleSelectors{
					Title: "h1",
					Body:  "article",
				},
			},
		})
		require.NoError(t, err)

		source := s.FindByName("test-source")
		require.NotNil(t, source)
		assert.Equal(t, "test-source", source.Name)
	})

	t.Run("update source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{
			{
				Name:      "test-source",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  2,
			},
		})

		updatedSource := &sourceutils.SourceConfig{
			Name:      "test-source",
			URL:       "https://updated.com",
			RateLimit: 2 * time.Second,
			MaxDepth:  3,
		}

		err := s.UpdateSource(t.Context(), updatedSource)
		require.NoError(t, err)

		source := s.FindByName("test-source")
		require.NotNil(t, source)
		assert.Equal(t, "https://updated.com", source.URL)
		assert.Equal(t, 2*time.Second, source.RateLimit)
		assert.Equal(t, 3, source.MaxDepth)
	})

	t.Run("delete source", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{
			{
				Name:      "test-source",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  2,
			},
		})

		err := s.DeleteSource(t.Context(), "test-source")
		require.NoError(t, err)

		source := s.FindByName("test-source")
		require.Nil(t, source)
	})
}

// TestMetrics tests source metrics
func TestMetrics(t *testing.T) {
	t.Parallel()

	t.Run("increment metrics", func(t *testing.T) {
		t.Parallel()
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{
			{
				Name:      "test-source",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  2,
				Selectors: sourceutils.SelectorConfig{
					Article: sourceutils.ArticleSelectors{
						Title: "h1",
						Body:  "article",
					},
				},
			},
		})

		metrics := s.GetMetrics()
		require.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.SourceCount)
		assert.NotZero(t, metrics.LastUpdated)
	})

	t.Run("metrics without sources", func(t *testing.T) {
		s := sourcetestutils.NewTestSources([]sourceutils.SourceConfig{})
		metrics := s.GetMetrics()
		require.Equal(t, int64(0), metrics.SourceCount)
		assert.NotZero(t, metrics.LastUpdated)
	})
}

// TestIndexNameHandling tests index name handling
func TestIndexNameHandling(t *testing.T) {
	t.Parallel()

	t.Run("default index name", func(t *testing.T) {
		t.Parallel()
		source := &sourceutils.SourceConfig{
			Name:      "Test Source",
			URL:       "https://test.com",
			RateLimit: time.Second,
			MaxDepth:  2,
		}
		s := sourcetestutils.NewTestSources(nil)
		err := s.AddSource(t.Context(), source)
		require.NoError(t, err)

		sources, err := s.ListSources(t.Context())
		require.NoError(t, err)
		require.Len(t, sources, 1)
		assert.Equal(t, "articles", sources[0].ArticleIndex)
		assert.Equal(t, "content", sources[0].Index)
	})

	t.Run("custom index names", func(t *testing.T) {
		t.Parallel()
		source := &sourceutils.SourceConfig{
			Name:         "Test Source",
			URL:          "https://test.com",
			RateLimit:    time.Second,
			MaxDepth:     2,
			ArticleIndex: "custom_articles",
			Index:        "custom_content",
		}
		s := sourcetestutils.NewTestSources(nil)
		err := s.AddSource(t.Context(), source)
		require.NoError(t, err)

		sources, err := s.ListSources(t.Context())
		require.NoError(t, err)
		require.Len(t, sources, 1)
		assert.Equal(t, "custom_articles", sources[0].ArticleIndex)
		assert.Equal(t, "custom_content", sources[0].Index)
	})
}

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	defaultConfig := sources.DefaultConfig()
	require.Equal(t, "articles", defaultConfig.ArticleIndex)
	require.Equal(t, "content", defaultConfig.Index)
}

// TestModule tests the sources module
func TestModule(t *testing.T) {
	t.Parallel()
	setup := newTestSetup(t)

	// Create mock logger
	mockLogger := testutils.NewMockLogger()

	app := fxtest.New(t,
		sources.Module,
		fx.Provide(
			func() config.Interface { return setup.config },
			func() logger.Interface { return mockLogger },
		),
		fx.Invoke(func(s sources.Interface) {
			require.NotNil(t, s)
			sources, err := s.ListSources(t.Context())
			require.NoError(t, err)
			require.NotNil(t, sources)
		}),
	)

	app.RequireStart()
	app.RequireStop()
}

// TestFindByName tests the FindByName method
func TestFindByName(t *testing.T) {
	t.Parallel()
	testConfigs := []sourceutils.SourceConfig{
		{
			Name:      "test1",
			URL:       "https://example1.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
		{
			Name:      "test2",
			URL:       "https://example2.com",
			RateLimit: 2 * time.Second,
			MaxDepth:  2,
		},
	}
	s := sourcetestutils.NewTestSources(testConfigs)
	require.NotNil(t, s)

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "existing source",
			source:  "test1",
			wantErr: false,
		},
		{
			name:    "non-existing source",
			source:  "test3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := s.FindByName(tt.source)
			if tt.wantErr {
				require.Nil(t, source)
				return
			}
			require.NotNil(t, source)
			require.Equal(t, tt.source, source.Name)
		})
	}
}
