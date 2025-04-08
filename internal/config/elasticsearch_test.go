package config_test

import (
	"testing"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(t *testing.T) *testutils.TestSetup
		wantErrMsg string
	}{
		{
			name: "valid config",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  tls:
    enabled: false
  retry:
    enabled: true
    initial_wait: 1s
    max_wait: 5s
    max_retries: 3
  bulk:
    size: 1000
    flush_interval: 30s
`, "")
			},
			wantErrMsg: "",
		},
		{
			name: "missing addresses",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses: []
  api_key: id:test_api_key
  index_name: test-index
`, "")
			},
			wantErrMsg: "elasticsearch addresses cannot be empty",
		},
		{
			name: "missing index name",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: ""
`, "")
			},
			wantErrMsg: "elasticsearch index name cannot be empty",
		},
		{
			name: "missing API key",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: ""
  index_name: test-index
`, "")
			},
			wantErrMsg: "elasticsearch API key cannot be empty",
		},
		{
			name: "invalid API key format",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: "invalid"
  index_name: test-index
`, "")
			},
			wantErrMsg: "elasticsearch API key must be in the format 'id:api_key'",
		},
		{
			name: "missing TLS certificate",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  tls:
    enabled: true
`, "")
			},
			wantErrMsg: "TLS certificate file is required when TLS is enabled",
		},
		{
			name: "invalid retry configuration",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  retry:
    enabled: true
    initial_wait: invalid
`, "")
			},
			wantErrMsg: "invalid retry configuration",
		},
		{
			name: "invalid bulk size",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  bulk:
    size: 0
`, "")
			},
			wantErrMsg: "bulk size must be greater than 0",
		},
		{
			name: "invalid flush interval",
			setup: func(t *testing.T) *testutils.TestSetup {
				return testutils.SetupTestEnvironment(t, `
app:
  environment: test
  name: gocrawl-test
  version: 0.0.1
crawler:
  base_url: http://test.example.com
  max_depth: 2
  parallelism: 2
  rate_limit: 2s
elasticsearch:
  addresses:
    - http://localhost:9200
  api_key: id:test_api_key
  index_name: test-index
  bulk:
    flush_interval: invalid
`, "")
			},
			wantErrMsg: "invalid flush interval",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setup := tt.setup(t)
			defer setup.Cleanup()

			cfg, err := config.LoadConfig(setup.ConfigPath)
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}
