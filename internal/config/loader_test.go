package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Ensure testdata directory exists
	err := os.MkdirAll("internal/config/testdata", 0755)
	require.NoError(t, err)

	// Create valid config file for testing
	validConfig := `
app:
  environment: development
  name: gocrawl
  version: 1.0.0
  debug: false
crawler:
  max_depth: 3
  parallelism: 2
  source_file: internal/config/testdata/sources.yml
log:
  level: debug
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: test_key
`
	err = os.WriteFile("internal/config/testdata/config.yml", []byte(validConfig), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/config.yml"); removeErr != nil {
			t.Errorf("Error removing test file: %v", removeErr)
		}
	}()

	// Create test sources file
	testSources := `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 1s
    max_depth: 2
    article_index: test_articles
    index: test_content
    selectors:
      article:
        title: h1
        content: article
        author: .author
        date: .date
`
	err = os.WriteFile("internal/config/testdata/sources.yml", []byte(testSources), 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/sources.yml"); removeErr != nil {
			t.Errorf("Error removing sources file: %v", removeErr)
		}
	}()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid config file",
			path:    "internal/config/testdata/config.yml",
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "internal/config/testdata/nonexistent.yml",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			path:    "internal/config/testdata/invalid.yml",
			wantErr: true,
		},
	}

	// Create invalid YAML file for testing
	invalidYAML := []byte("invalid: yaml: content")
	err = os.WriteFile("internal/config/testdata/invalid.yml", invalidYAML, 0644)
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove("internal/config/testdata/invalid.yml"); removeErr != nil {
			t.Errorf("Error removing test file: %v", removeErr)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, loadErr := config.LoadConfig(tt.path)
			if tt.wantErr {
				require.Error(t, loadErr)
			} else {
				require.NoError(t, loadErr)
			}
		})
	}
}

func TestDefaultArticleSelectors(t *testing.T) {
	selectors := config.DefaultArticleSelectors()

	// Test that default selectors are not empty
	require.NotEmpty(t, selectors.Container)
	require.NotEmpty(t, selectors.Title)
	require.NotEmpty(t, selectors.Body)
	require.NotEmpty(t, selectors.Intro)
	require.NotEmpty(t, selectors.Byline)
	require.NotEmpty(t, selectors.PublishedTime)
	require.NotEmpty(t, selectors.TimeAgo)
	require.NotEmpty(t, selectors.JSONLD)
	require.NotEmpty(t, selectors.Section)
	require.NotEmpty(t, selectors.Keywords)
	require.NotEmpty(t, selectors.Description)
	require.NotEmpty(t, selectors.OGTitle)
	require.NotEmpty(t, selectors.OGDescription)
	require.NotEmpty(t, selectors.OGImage)
	require.NotEmpty(t, selectors.OgURL)
	require.NotEmpty(t, selectors.Canonical)

	// Test that selectors have meaningful values
	require.Equal(t, "article, .article, [itemtype*='Article']", selectors.Container)
	require.Equal(t, "h1", selectors.Title)
	require.Equal(t, "article, [role='main'], .content, .article-content", selectors.Body)
	require.Equal(t, ".article-intro, .post-intro, .entry-summary", selectors.Intro)
	require.Equal(t, ".article-byline, .post-meta, .entry-meta", selectors.Byline)
	require.Equal(t, "meta[property='article:published_time']", selectors.PublishedTime)
	require.Equal(t, "time", selectors.TimeAgo)
	require.Equal(t, "script[type='application/ld+json']", selectors.JSONLD)
	require.Equal(t, "meta[property='article:section']", selectors.Section)
	require.Equal(t, "meta[name='keywords']", selectors.Keywords)
	require.Equal(t, "meta[name='description']", selectors.Description)
	require.Equal(t, "meta[property='og:title']", selectors.OGTitle)
	require.Equal(t, "meta[property='og:description']", selectors.OGDescription)
	require.Equal(t, "meta[property='og:image']", selectors.OGImage)
	require.Equal(t, "meta[property='og:url']", selectors.OgURL)
	require.Equal(t, "link[rel='canonical']", selectors.Canonical)
}
