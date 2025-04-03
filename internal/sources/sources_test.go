// Package sources_test provides tests for the sources package.
package sources_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/jonesrussell/gocrawl/internal/sourceutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary sources.yml file
	tmpDir := t.TempDir()
	sourcesFile := filepath.Join(tmpDir, "sources.yml")
	err := os.WriteFile(sourcesFile, []byte(`
sources:
  - name: test-source
    url: https://example.com
    rate_limit: 1s
    max_depth: 2
    selectors:
      article:
        title: h1
        body: article
`), 0644)
	require.NoError(t, err)

	// Set environment variables
	t.Setenv("SOURCES_FILE", sourcesFile)
	defer os.Unsetenv("SOURCES_FILE")

	// Load sources from file
	loaderConfigs, err := loader.LoadFromFile(sourcesFile)
	require.NoError(t, err)
	require.Len(t, loaderConfigs, 1)

	// Convert loader.Config to sourceutils.SourceConfig
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

	// Create test sources instance
	s := testutils.NewTestSources(sourceConfigs)
	require.NotNil(t, s)

	// Test ListSources
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "test-source", sources[0].Name)
	assert.Equal(t, "https://example.com", sources[0].URL)
	assert.Equal(t, time.Second, sources[0].RateLimit)
	assert.Equal(t, 2, sources[0].MaxDepth)
	assert.Equal(t, "h1", sources[0].Selectors.Article.Title)
	assert.Equal(t, "article", sources[0].Selectors.Article.Body)
}

func TestGetSource(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{
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
	require.NotNil(t, s)

	// Test FindByName
	source, err := s.FindByName("test-source")
	require.NoError(t, err)
	require.NotNil(t, source)
	assert.Equal(t, "test-source", source.Name)
	assert.Equal(t, "https://example.com", source.URL)
	assert.Equal(t, time.Second, source.RateLimit)
	assert.Equal(t, 2, source.MaxDepth)
	assert.Equal(t, "h1", source.Selectors.Article.Title)
	assert.Equal(t, "article", source.Selectors.Article.Body)
}

func TestValidateSource(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{})
	require.NotNil(t, s)

	// Test ValidateSource with valid source
	err := s.ValidateSource(&sourceutils.SourceConfig{
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

	// Test ValidateSource with invalid source
	err = s.ValidateSource(&sourceutils.SourceConfig{
		Name:      "",
		URL:       "",
		RateLimit: 0,
		MaxDepth:  0,
	})
	require.NoError(t, err)
}

func TestAddSource(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{})
	require.NotNil(t, s)

	// Test AddSource
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

	// Verify source was added
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "test-source", sources[0].Name)
}

func TestUpdateSource(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{
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
	require.NotNil(t, s)

	// Test UpdateSource
	err := s.UpdateSource(t.Context(), &sourceutils.SourceConfig{
		Name:      "test-source",
		URL:       "https://updated.example.com",
		RateLimit: 2 * time.Second,
		MaxDepth:  3,
		Selectors: sourceutils.SelectorConfig{
			Article: sourceutils.ArticleSelectors{
				Title: "h2",
				Body:  "div.article",
			},
		},
	})
	require.NoError(t, err)

	// Verify source was updated
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	assert.Equal(t, "https://updated.example.com", sources[0].URL)
	assert.Equal(t, 2*time.Second, sources[0].RateLimit)
	assert.Equal(t, 3, sources[0].MaxDepth)
	assert.Equal(t, "h2", sources[0].Selectors.Article.Title)
	assert.Equal(t, "div.article", sources[0].Selectors.Article.Body)
}

func TestDeleteSource(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{
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
	require.NotNil(t, s)

	// Test DeleteSource
	err := s.DeleteSource(t.Context(), "test-source")
	require.NoError(t, err)

	// Verify source was deleted
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Empty(t, sources)
}

func TestGetMetrics(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{
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
	require.NotNil(t, s)

	// Test GetMetrics
	metrics := s.GetMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.SourceCount)
	assert.NotZero(t, metrics.LastUpdated)
}

func TestFindByName(t *testing.T) {
	// Create test sources instance
	s := testutils.NewTestSources([]sourceutils.SourceConfig{
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
	require.NotNil(t, s)

	// Test FindByName with existing source
	source, err := s.FindByName("test-source")
	require.NoError(t, err)
	require.NotNil(t, source)
	assert.Equal(t, "test-source", source.Name)

	// Test FindByName with non-existent source
	source, err = s.FindByName("non-existent")
	require.NoError(t, err)
	require.Nil(t, source)
}

func TestIndexNameHandling(t *testing.T) {
	// Create a test sources instance
	s := testutils.NewTestSources(nil)
	require.NotNil(t, s)

	// Test source with empty index names
	source := &sourceutils.SourceConfig{
		Name:      "Test Source",
		URL:       "https://test.com",
		RateLimit: time.Second,
		MaxDepth:  2,
	}
	err := s.AddSource(t.Context(), source)
	require.NoError(t, err)

	// Verify default index names were set
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	require.Equal(t, "articles", sources[0].ArticleIndex)
	require.Equal(t, "content", sources[0].Index)
}

func TestDefaultConfigIndexNames(t *testing.T) {
	// Test DefaultConfig
	defaultConfig := sources.DefaultConfig()
	require.Equal(t, "articles", defaultConfig.ArticleIndex)
	require.Equal(t, "content", defaultConfig.Index)
}

func TestSourceIndexNamePersistence(t *testing.T) {
	// Create a test sources instance
	s := testutils.NewTestSources(nil)
	require.NotNil(t, s)

	// Test source with custom index names
	source := &sourceutils.SourceConfig{
		Name:         "Test Source",
		URL:          "https://test.com",
		RateLimit:    time.Second,
		MaxDepth:     2,
		ArticleIndex: "custom_articles",
		Index:        "custom_content",
	}
	err := s.AddSource(t.Context(), source)
	require.NoError(t, err)

	// Verify custom index names were set
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	require.Equal(t, "custom_articles", sources[0].ArticleIndex)
	require.Equal(t, "custom_content", sources[0].Index)

	// Test updating source with custom index names
	updatedSource := &sourceutils.SourceConfig{
		Name:         "Test Source",
		URL:          "https://updated.com",
		RateLimit:    2 * time.Second,
		MaxDepth:     3,
		ArticleIndex: "updated_articles",
		Index:        "updated_content",
	}
	err = s.UpdateSource(t.Context(), updatedSource)
	require.NoError(t, err)

	// Verify index names were updated
	sources, err = s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	require.Equal(t, "updated_articles", sources[0].ArticleIndex)
	require.Equal(t, "updated_content", sources[0].Index)
}

func TestProvideSourcesIndexNames(t *testing.T) {
	// Create a test sources instance
	s := testutils.NewTestSources(nil)
	require.NotNil(t, s)

	// Test source with custom index names
	source := &sourceutils.SourceConfig{
		Name:         "Test Source",
		URL:          "https://test.com",
		RateLimit:    time.Second,
		MaxDepth:     2,
		ArticleIndex: "custom_articles",
		Index:        "custom_content",
	}
	err := s.AddSource(t.Context(), source)
	require.NoError(t, err)

	// Verify custom index names were set
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	require.Equal(t, "custom_articles", sources[0].ArticleIndex)
	require.Equal(t, "custom_content", sources[0].Index)

	// Test updating source with custom index names
	updatedSource := &sourceutils.SourceConfig{
		Name:         "Test Source",
		URL:          "https://updated.com",
		RateLimit:    2 * time.Second,
		MaxDepth:     3,
		ArticleIndex: "updated_articles",
		Index:        "updated_content",
	}
	err = s.UpdateSource(t.Context(), updatedSource)
	require.NoError(t, err)

	// Verify index names were updated
	sources, err = s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, sources, 1)
	require.Equal(t, "updated_articles", sources[0].ArticleIndex)
	require.Equal(t, "updated_content", sources[0].Index)
}

func TestEmptySources(t *testing.T) {
	// Create a test sources instance with no sources
	s := testutils.NewTestSources(nil)
	require.NotNil(t, s)

	// Test ListSources with empty sources
	sources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Empty(t, sources)

	// Test GetSources with empty sources
	configs, err := s.GetSources()
	require.NoError(t, err)
	require.Empty(t, configs)

	// Test GetMetrics with empty sources
	metrics := s.GetMetrics()
	require.Equal(t, int64(0), metrics.SourceCount)
	assert.NotZero(t, metrics.LastUpdated)
}
