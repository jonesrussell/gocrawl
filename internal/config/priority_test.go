package config_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/stretchr/testify/require"
)

func TestPriorityConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		sources     string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid priority configuration",
			config: map[string]interface{}{
				"app.environment":          "test",
				"app.name":                 "gocrawl-test",
				"app.version":              "0.0.1",
				"log.level":                "debug",
				"elasticsearch.addresses":  []string{"http://localhost:9200"},
				"elasticsearch.api_key":    "id:test_api_key",
				"elasticsearch.index_name": "test-index",
				"crawler.base_url":         "http://test.example.com",
				"crawler.max_depth":        2,
				"crawler.rate_limit":       "2s",
				"crawler.parallelism":      2,
			},
			sources: `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    article_index: test_articles
    index: test_content
    priority: 1
    selectors:
      article:
        title: h1
        body: article
        author: .author
        published_time: .date
`,
			wantErr: false,
		},
		{
			name: "invalid priority pattern",
			config: map[string]interface{}{
				"app.environment":          "test",
				"app.name":                 "gocrawl-test",
				"app.version":              "0.0.1",
				"log.level":                "debug",
				"elasticsearch.addresses":  []string{"http://localhost:9200"},
				"elasticsearch.api_key":    "id:test_api_key",
				"elasticsearch.index_name": "test-index",
				"crawler.base_url":         "http://test.example.com",
				"crawler.max_depth":        2,
				"crawler.rate_limit":       "2s",
				"crawler.parallelism":      2,
			},
			sources: `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    article_index: test_articles
    index: test_content
    priority: invalid
    selectors:
      article:
        title: h1
        body: article
        author: .author
        published_time: .date
`,
			wantErr:     true,
			errContains: "invalid priority pattern",
		},
		{
			name: "invalid priority value",
			config: map[string]interface{}{
				"app.environment":          "test",
				"app.name":                 "gocrawl-test",
				"app.version":              "0.0.1",
				"log.level":                "debug",
				"elasticsearch.addresses":  []string{"http://localhost:9200"},
				"elasticsearch.api_key":    "id:test_api_key",
				"elasticsearch.index_name": "test-index",
				"crawler.base_url":         "http://test.example.com",
				"crawler.max_depth":        2,
				"crawler.rate_limit":       "2s",
				"crawler.parallelism":      2,
			},
			sources: `
sources:
  - name: test_source
    url: http://test.example.com
    rate_limit: 2s
    max_depth: 2
    article_index: test_articles
    index: test_content
    priority: -1
    selectors:
      article:
        title: h1
        body: article
        author: .author
        published_time: .date
`,
			wantErr:     true,
			errContains: "priority must be greater than or equal to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := testutils.SetupTestEnvironment(t, tt.config, tt.sources)
			defer setup.Cleanup()

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
