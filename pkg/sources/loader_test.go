package sources_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/pkg/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, configs []sources.Config)
	}{
		{
			name: "valid config",
			content: `
- name: test1
  url: http://example.com
  rate_limit: 1s
  max_depth: 2
  time: ["00:00", "12:00"]
- name: test2
  url: https://example.org
  rate_limit: 2s
  max_depth: 3
  time: ["03:00"]
`,
			wantErr: false,
			validate: func(t *testing.T, configs []sources.Config) {
				require.Len(t, configs, 2)

				// Check first config
				assert.Equal(t, "test1", configs[0].Name)
				assert.Equal(t, "http://example.com", configs[0].URL)
				assert.Equal(t, time.Second, configs[0].RateLimit)
				assert.Equal(t, 2, configs[0].MaxDepth)
				assert.Equal(t, []string{"00:00", "12:00"}, configs[0].Time)

				// Check second config
				assert.Equal(t, "test2", configs[1].Name)
				assert.Equal(t, "https://example.org", configs[1].URL)
				assert.Equal(t, 2*time.Second, configs[1].RateLimit)
				assert.Equal(t, 3, configs[1].MaxDepth)
				assert.Equal(t, []string{"03:00"}, configs[1].Time)
			},
		},
		{
			name: "invalid YAML",
			content: `
- name: test
  url: http://example.com
  rate_limit: invalid
`,
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
		{
			name: "missing required fields",
			content: `
- name: test1
  url: http://example.com
  rate_limit: 1s
  max_depth: 2
- name: test2
  rate_limit: 2s
  max_depth: 3
`,
			wantErr: true,
			errMsg:  "invalid config at index 1",
		},
		{
			name: "invalid URL",
			content: `
- name: test
  url: not-a-url
  rate_limit: 1s
  max_depth: 2
`,
			wantErr: true,
			errMsg:  "invalid URL",
		},
		{
			name: "negative rate limit",
			content: `
- name: test
  url: http://example.com
  rate_limit: -1s
  max_depth: 2
`,
			wantErr: true,
			errMsg:  "rate limit must be positive",
		},
		{
			name: "negative max depth",
			content: `
- name: test
  url: http://example.com
  rate_limit: 1s
  max_depth: -2
`,
			wantErr: true,
			errMsg:  "max depth must be positive",
		},
		{
			name: "invalid time format",
			content: `
- name: test
  url: http://example.com
  rate_limit: 1s
  max_depth: 2
  time: ["25:00"]
`,
			wantErr: true,
			errMsg:  "invalid time format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the test content
			tmpFile := filepath.Join(tmpDir, "test.yaml")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Try to load the config
			configs, err := sources.LoadFromFile(tmpFile)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, configs)
			}
		})
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	_, err := sources.LoadFromFile("non-existent.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}
