// Package sources_test provides tests for the sources package.
package sources_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	t.Parallel()
	// Create a temporary sources.yml file for testing
	tmpDir := t.TempDir()
	sourcesYml := `sources:
  - name: Test Source
    url: https://test.com
    rate_limit: 1s
    max_depth: 2
    index: test_content
    article_index: test_articles
`
	err := os.WriteFile(filepath.Join(tmpDir, "sources.yml"), []byte(sourcesYml), 0644)
	require.NoError(t, err)

	// Set environment variables for testing
	t.Setenv("SOURCES_FILE", filepath.Join(tmpDir, "sources.yml"))
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")

	// Create a new Sources instance
	s := testutils.NewTestInterface(nil)
	require.NotNil(t, s)

	// Load sources from file
	allSources, err := s.ListSources(t.Context())
	require.NoError(t, err)
	require.Len(t, allSources, 1)

	// Verify source details
	source := allSources[0]
	require.Equal(t, "Test Source", source.Name)
	require.Equal(t, "https://test.com", source.URL)
	require.Equal(t, time.Second, source.RateLimit)
	require.Equal(t, 2, source.MaxDepth)
	require.Equal(t, "test_content", source.Index)
	require.Equal(t, "test_articles", source.ArticleIndex)
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
