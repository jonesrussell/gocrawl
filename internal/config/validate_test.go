package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/config/testutils"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*testing.T)
		expectedError string
	}{
		{
			name: "valid configuration",
			setup: func(t *testing.T) {
				// Set required environment variables
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "",
		},
		{
			name: "invalid environment",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "invalid")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "invalid environment: invalid",
		},
		{
			name: "invalid log level",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "invalid")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "invalid log level: invalid",
		},
		{
			name: "invalid crawler max depth",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "0")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "crawler max depth must be greater than 0",
		},
		{
			name: "invalid crawler parallelism",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "0")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "false")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "crawler parallelism must be greater than 0",
		},
		{
			name: "server security enabled without API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "")
			},
			expectedError: "server security is enabled but no API key is provided",
		},
		{
			name: "server security enabled with invalid API key",
			setup: func(t *testing.T) {
				t.Setenv("GOCRAWL_APP_ENVIRONMENT", "test")
				t.Setenv("GOCRAWL_LOG_LEVEL", "debug")
				t.Setenv("GOCRAWL_CRAWLER_MAX_DEPTH", "2")
				t.Setenv("GOCRAWL_CRAWLER_PARALLELISM", "2")
				t.Setenv("GOCRAWL_SERVER_SECURITY_ENABLED", "true")
				t.Setenv("GOCRAWL_SERVER_SECURITY_API_KEY", "invalid")
			},
			expectedError: "invalid API key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := testutils.SetupTestEnv(t)
			defer cleanup()

			// Run test setup
			tt.setup(t)

			// Create config
			cfg, err := config.NewConfig(testutils.NewTestLogger(t))

			// Validate results
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
		})
	}
}

func TestElasticsearchConfigBasicValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      config.ElasticsearchConfig
		env         map[string]string
		expectedErr string
	}{
		{
			name: "valid config",
			config: config.ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test-index",
				APIKey:    "id:api_key",
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "",
		},
		{
			name: "missing addresses",
			config: config.ElasticsearchConfig{
				IndexName: "test-index",
				APIKey:    "id:api_key",
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "elasticsearch addresses cannot be empty",
		},
		{
			name: "missing index name",
			config: config.ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "id:api_key",
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "elasticsearch index name cannot be empty",
		},
		{
			name: "missing API key",
			config: config.ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test-index",
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "elasticsearch API key cannot be empty",
		},
		{
			name: "invalid API key format",
			config: config.ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test-index",
				APIKey:    "invalid-key",
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "elasticsearch API key must be in the format 'id:api_key'",
		},
		{
			name: "missing TLS certificate",
			config: config.ElasticsearchConfig{
				Addresses: []string{"https://localhost:9200"},
				IndexName: "test-index",
				APIKey:    "id:api_key",
				TLS: config.TLSConfig{
					Enabled: true,
				},
			},
			env: map[string]string{
				"GOCRAWL_APP_ENVIRONMENT":     "test",
				"GOCRAWL_APP_NAME":            "gocrawl-test",
				"GOCRAWL_APP_VERSION":         "0.0.1",
				"GOCRAWL_CRAWLER_BASE_URL":    "http://test.example.com",
				"GOCRAWL_CRAWLER_MAX_DEPTH":   "2",
				"GOCRAWL_CRAWLER_RATE_LIMIT":  "2s",
				"GOCRAWL_CRAWLER_PARALLELISM": "2",
			},
			expectedErr: "TLS certificate file is required when TLS is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			cleanup := testutils.SetupTestEnv(t)
			defer cleanup()

			// Set environment variables
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			// Parse rate limit duration
			rateLimit, err := time.ParseDuration(tt.env["GOCRAWL_CRAWLER_RATE_LIMIT"])
			require.NoError(t, err)

			// Create test config
			cfg := &config.Config{
				App: config.AppConfig{
					Environment: tt.env["GOCRAWL_APP_ENVIRONMENT"],
					Name:        tt.env["GOCRAWL_APP_NAME"],
					Version:     tt.env["GOCRAWL_APP_VERSION"],
				},
				Crawler: config.CrawlerConfig{
					BaseURL:     tt.env["GOCRAWL_CRAWLER_BASE_URL"],
					MaxDepth:    2,
					RateLimit:   rateLimit,
					Parallelism: 2,
				},
				Elasticsearch: tt.config,
			}

			// Validate config
			err = config.ValidateConfig(cfg)

			if tt.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			}
		})
	}
}
