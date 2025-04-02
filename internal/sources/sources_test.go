// Package sources_test provides tests for the sources package.
package sources_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	configtestutils "github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/loader"
	"github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary sources.yml file for testing
	tmpDir := t.TempDir()
	sourcesYml := `sources:
  - name: Test Source
    url: https://test.com
    rate_limit: 1s
    max_depth: 2
`
	writeErr := os.WriteFile(filepath.Join(tmpDir, "sources.yml"), []byte(sourcesYml), 0644)
	require.NoError(t, writeErr)

	// Set environment variables for testing
	t.Setenv("SOURCES_FILE", filepath.Join(tmpDir, "sources.yml"))
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")

	// Load sources from file
	loaderConfigs, err := loader.LoadFromFile(filepath.Join(tmpDir, "sources.yml"))
	require.NoError(t, err)
	require.Len(t, loaderConfigs, 1)

	// Convert loader.Config to sources.Config
	var configs []sources.Config
	for _, src := range loaderConfigs {
		rateLimit, parseErr := time.ParseDuration(src.RateLimit)
		require.NoError(t, parseErr)

		configs = append(configs, sources.Config{
			Name:      src.Name,
			URL:       src.URL,
			RateLimit: rateLimit,
			MaxDepth:  src.MaxDepth,
			Time:      src.Time,
		})
	}

	// Create a new Sources instance with the loaded configs
	s := testutils.NewTestInterface(configs)
	require.NotNil(t, s)

	// Get all sources
	allSources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, allSources, 1)

	// Verify source details
	source := allSources[0]
	require.Equal(t, "Test Source", source.Name)
	require.Equal(t, "https://test.com", source.URL)
	require.Equal(t, time.Second, source.RateLimit)
	require.Equal(t, 2, source.MaxDepth)
}

func TestGetSource(t *testing.T) {
	t.Parallel()
	testConfigs := []sources.Config{
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
	s := testutils.NewTestInterface(testConfigs)
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
			source, err := s.GetSource(t.Context(), tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.source, source.Name)
		})
	}
}

func TestValidateSource(t *testing.T) {
	testConfigs := []sources.Config{
		{
			Name:      "test",
			URL:       "https://example.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
	}
	s := testutils.NewTestInterface(testConfigs)
	require.NotNil(t, s)

	tests := []struct {
		name    string
		source  *sources.Config
		wantErr bool
	}{
		{
			name: "valid source",
			source: &sources.Config{
				Name:      "test",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: false,
		},
		{
			name:    "nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name: "missing name",
			source: &sources.Config{
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: true,
		},
		{
			name: "missing URL",
			source: &sources.Config{
				Name:      "test",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: true,
		},
		{
			name: "missing rate limit",
			source: &sources.Config{
				Name:     "test",
				URL:      "https://example.com",
				MaxDepth: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid max depth",
			source: &sources.Config{
				Name:      "test",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateErr := s.ValidateSource(tt.source)
			if tt.wantErr {
				require.Error(t, validateErr)
				return
			}
			require.NoError(t, validateErr)
		})
	}
}

func TestAddSource(t *testing.T) {
	t.Parallel()
	s := testutils.NewTestInterface(nil)
	require.NotNil(t, s)

	tests := []struct {
		name    string
		source  *sources.Config
		wantErr bool
	}{
		{
			name: "valid source",
			source: &sources.Config{
				Name:      "test1",
				URL:       "https://example1.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: false,
		},
		{
			name: "duplicate source",
			source: &sources.Config{
				Name:      "test1",
				URL:       "https://example1.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := s.AddSource(t.Context(), tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify source was added
			source, err := s.GetSource(t.Context(), tt.source.Name)
			require.NoError(t, err)
			require.Equal(t, tt.source.Name, source.Name)
		})
	}
}

func TestUpdateSource(t *testing.T) {
	t.Parallel()
	testConfigs := []sources.Config{
		{
			Name:      "test1",
			URL:       "https://example1.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
	}
	s := testutils.NewTestInterface(testConfigs)
	require.NotNil(t, s)

	tests := []struct {
		name    string
		source  *sources.Config
		wantErr bool
	}{
		{
			name: "existing source",
			source: &sources.Config{
				Name:      "test1",
				URL:       "https://example1.com/updated",
				RateLimit: 2 * time.Second,
				MaxDepth:  2,
			},
			wantErr: false,
		},
		{
			name: "non-existing source",
			source: &sources.Config{
				Name:      "test2",
				URL:       "https://example2.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := s.UpdateSource(t.Context(), tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify source was updated
			source, err := s.GetSource(t.Context(), tt.source.Name)
			require.NoError(t, err)
			require.Equal(t, tt.source.URL, source.URL)
			require.Equal(t, tt.source.RateLimit, source.RateLimit)
			require.Equal(t, tt.source.MaxDepth, source.MaxDepth)
		})
	}
}

func TestDeleteSource(t *testing.T) {
	t.Parallel()
	testConfigs := []sources.Config{
		{
			Name:      "test1",
			URL:       "https://example1.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
	}
	s := testutils.NewTestInterface(testConfigs)
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
			source:  "test2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := s.DeleteSource(t.Context(), tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify source was deleted
			_, err = s.GetSource(t.Context(), tt.source)
			require.Error(t, err)
		})
	}
}

func TestGetMetrics(t *testing.T) {
	testConfigs := []sources.Config{
		{
			Name:      "test",
			URL:       "https://example.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
	}
	s := testutils.NewTestInterface(testConfigs)
	require.NotNil(t, s)

	metrics := s.GetMetrics()
	require.Equal(t, int64(1), metrics.SourceCount)
}

func TestFindByName(t *testing.T) {
	testConfigs := []sources.Config{
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
	s := testutils.NewTestInterface(testConfigs)
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
			source, err := s.FindByName(tt.source)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.source, source.Name)
		})
	}
}

func TestIndexNameHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sourceConfig   sources.Config
		wantArticleIdx string
		wantContentIdx string
	}{
		{
			name: "default config",
			sourceConfig: sources.Config{
				Name:      "test",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantArticleIdx: "articles",
			wantContentIdx: "content",
		},
		{
			name: "custom index names",
			sourceConfig: sources.Config{
				Name:         "test",
				URL:          "https://example.com",
				RateLimit:    time.Second,
				MaxDepth:     1,
				ArticleIndex: "custom_articles",
				Index:        "custom_content",
			},
			wantArticleIdx: "custom_articles",
			wantContentIdx: "custom_content",
		},
		{
			name: "empty index names",
			sourceConfig: sources.Config{
				Name:      "test",
				URL:       "https://example.com",
				RateLimit: time.Second,
				MaxDepth:  1,
			},
			wantArticleIdx: "articles",
			wantContentIdx: "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a new Sources instance with the test config
			s := testutils.NewTestInterface([]sources.Config{tt.sourceConfig})
			require.NotNil(t, s)

			// Get the source
			source, err := s.GetSource(t.Context(), tt.sourceConfig.Name)
			require.NoError(t, err)
			require.NotNil(t, source)

			// Verify index names
			require.Equal(t, tt.wantArticleIdx, source.ArticleIndex, "ArticleIndex mismatch")
			require.Equal(t, tt.wantContentIdx, source.Index, "Index mismatch")
		})
	}
}

func TestDefaultConfigIndexNames(t *testing.T) {
	t.Parallel()

	// Test NewConfig
	newConfig := sources.Config{}
	require.Equal(t, "articles", newConfig.ArticleIndex, "NewConfig ArticleIndex mismatch")
	require.Equal(t, "content", newConfig.Index, "NewConfig Index mismatch")

	// Test DefaultConfig
	defaultConfig := sources.Config{}
	require.Equal(t, "articles", defaultConfig.ArticleIndex, "DefaultConfig ArticleIndex mismatch")
	require.Equal(t, "content", defaultConfig.Index, "DefaultConfig Index mismatch")
}

func TestSourceIndexNamePersistence(t *testing.T) {
	t.Parallel()

	// Create a source with custom index names
	sourceConfig := sources.Config{
		Name:         "test",
		URL:          "https://example.com",
		RateLimit:    time.Second,
		MaxDepth:     1,
		ArticleIndex: "custom_articles",
		Index:        "custom_content",
	}

	// Create a new Sources instance
	s := testutils.NewTestInterface([]sources.Config{sourceConfig})
	require.NotNil(t, s)

	// Add the source
	err := s.AddSource(t.Context(), &sourceConfig)
	require.NoError(t, err)

	// Get the source back
	source, err := s.GetSource(t.Context(), sourceConfig.Name)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Verify index names persisted
	require.Equal(t, "custom_articles", source.ArticleIndex, "ArticleIndex not persisted")
	require.Equal(t, "custom_content", source.Index, "Index not persisted")

	// Update the source
	source.ArticleIndex = "updated_articles"
	source.Index = "updated_content"
	err = s.UpdateSource(t.Context(), source)
	require.NoError(t, err)

	// Get the source again
	source, err = s.GetSource(t.Context(), sourceConfig.Name)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Verify index names were updated
	require.Equal(t, "updated_articles", source.ArticleIndex, "ArticleIndex not updated")
	require.Equal(t, "updated_content", source.Index, "Index not updated")
}

func TestProvideSourcesIndexNames(t *testing.T) {
	t.Parallel()

	// Create a mock config with sources
	mockConfig := configtestutils.NewMockConfig().WithSources([]config.Source{
		{
			Name:         "test1",
			URL:          "https://example1.com",
			RateLimit:    time.Second,
			MaxDepth:     1,
			ArticleIndex: "custom_articles_1",
			Index:        "custom_content_1",
		},
		{
			Name:      "test2",
			URL:       "https://example2.com",
			RateLimit: 2 * time.Second,
			MaxDepth:  2,
		},
	})

	// Create test module params
	params := sources.ModuleParams{
		Config: mockConfig,
		Logger: logger.NewNoOp(),
	}

	// Create sources instance
	s := sources.ProvideSources(params)
	require.NotNil(t, s)

	// Test source with custom index names
	source1, err := s.GetSource(t.Context(), "test1")
	require.NoError(t, err)
	require.NotNil(t, source1)
	require.Equal(t, "custom_articles_1", source1.ArticleIndex, "Custom ArticleIndex not loaded")
	require.Equal(t, "custom_content_1", source1.Index, "Custom Index not loaded")

	// Test source with default index names
	source2, err := s.GetSource(t.Context(), "test2")
	require.NoError(t, err)
	require.NotNil(t, source2)
	require.Equal(t, "articles", source2.ArticleIndex, "Default ArticleIndex not set")
	require.Equal(t, "content", source2.Index, "Default Index not set")

	// Test empty sources case
	emptyConfig := configtestutils.NewMockConfig().WithSources([]config.Source{})
	emptyParams := sources.ModuleParams{
		Config: emptyConfig,
		Logger: logger.NewNoOp(),
	}
	emptySources := sources.ProvideSources(emptyParams)
	require.NotNil(t, emptySources)

	// Verify default source has correct index names
	defaultSource, err := emptySources.GetSource(t.Context(), "default")
	require.NoError(t, err)
	require.NotNil(t, defaultSource)
	require.Equal(t, "articles", defaultSource.ArticleIndex, "Default source ArticleIndex mismatch")
	require.Equal(t, "content", defaultSource.Index, "Default source Index mismatch")
}
