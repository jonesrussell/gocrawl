package config_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/stretchr/testify/require"
)

func TestPriorityConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configContent  string
		sourcesContent string
		wantErr        bool
		errMsg         string
	}{
		{
			name: "valid configuration",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 2
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: false,
		},
		{
			name: "invalid max depth",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 0
  rate_limit: 2s
  parallelism: 2
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: true,
			errMsg:  "max depth must be greater than 0",
		},
		{
			name: "invalid parallelism",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: 2s
  parallelism: 0
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: true,
			errMsg:  "parallelism must be greater than 0",
		},
		{
			name: "invalid rate limit",
			configContent: `
crawler:
  base_url: http://test.example.com
  max_depth: 2
  rate_limit: invalid
  parallelism: 2
  source_file: sources.yml
`,
			sourcesContent: `sources:
  - name: test
    url: http://test.example.com
    selectors:
      - name: title
        path: h1`,
			wantErr: true,
			errMsg:  "invalid rate limit duration",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setup := testutils.SetupTestEnvironment(t, tt.configContent, tt.sourcesContent)
			defer setup.Cleanup()

			cfg, err := config.New(testutils.NewTestLogger(t))
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Validate specific fields
			crawlerCfg := cfg.GetCrawlerConfig()
			require.Equal(t, "http://test.example.com", crawlerCfg.BaseURL)
			require.Equal(t, 2, crawlerCfg.MaxDepth)
			require.Equal(t, 2*time.Second, crawlerCfg.RateLimit)
			require.Equal(t, 2, crawlerCfg.Parallelism)
			require.Equal(t, setup.SourcesPath, crawlerCfg.SourceFile)
		})
	}
}
