// Package sources_test provides tests for the sources package.
package sources_test

import (
	"os"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/jonesrussell/gocrawl/internal/sources/testutils"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary file with test data
	content := `sources:
  - name: test_source
    url: https://example.com
    rate_limit: 1s
    max_depth: 2
    article_index: articles
    index: content
    time:
      - published_time
      - time_ago
    selectors:
      article:
        container: article
        title: h1
        body: .content
        intro: .intro
        byline: .byline
        published_time: time
        time_ago: .time-ago
        jsonld: script[type="application/ld+json"]
        section: .section
        keywords: meta[name=keywords]
        description: meta[name=description]
        og_title: meta[property="og:title"]
        og_description: meta[property="og:description"]
        og_image: meta[property="og:image"]
        og_url: meta[property="og:url"]
        canonical: link[rel=canonical]
        word_count: .word-count
        publish_date: time[pubdate]
        category: .category
        tags: .tags
        author: .author
        byline_name: .byline-name`

	tmpfile, err := os.CreateTemp(t.TempDir(), "sources_test")
	require.NoError(t, err)
	defer func(name string) {
		if removeErr := os.Remove(name); removeErr != nil {
			t.Errorf("Error removing test file: %v", removeErr)
		}
	}(tmpfile.Name())

	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	s, err := sources.LoadFromFile(tmpfile.Name())
	require.NoError(t, err)
	require.NotNil(t, s)

	allSources := s.GetSources()
	require.Len(t, allSources, 1)
	require.Equal(t, "test_source", allSources[0].Name)
	require.Equal(t, "https://example.com", allSources[0].URL)
	require.Equal(t, time.Second, allSources[0].RateLimit)
	require.Equal(t, 2, allSources[0].MaxDepth)
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
	s := testutils.NewTestSources(testConfigs)
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

func TestValidate(t *testing.T) {
	testConfigs := []sources.Config{
		{
			Name:      "test",
			URL:       "https://example.com",
			RateLimit: time.Second,
			MaxDepth:  1,
		},
	}
	s := testutils.NewTestSources(testConfigs)
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
			validateErr := s.Validate(tt.source)
			if tt.wantErr {
				require.Error(t, validateErr)
				return
			}
			require.NoError(t, validateErr)
		})
	}
}
