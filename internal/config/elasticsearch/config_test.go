package elasticsearch_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/stretchr/testify/require"
)

// testConfig returns a valid base configuration for testing
func testConfig() *elasticsearch.Config {
	return &elasticsearch.Config{
		Addresses:     []string{"http://localhost:9200"},
		IndexName:     "test-index",
		APIKey:        "valid_id:valid_key",
		BulkSize:      1000,
		FlushInterval: 30 * time.Second,
		Retry: struct {
			Enabled     bool          `yaml:"enabled"`
			InitialWait time.Duration `yaml:"initial_wait"`
			MaxWait     time.Duration `yaml:"max_wait"`
			MaxRetries  int           `yaml:"max_retries"`
		}{
			Enabled:     true,
			InitialWait: 1 * time.Second,
			MaxWait:     5 * time.Second,
			MaxRetries:  3,
		},
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modify      func(*elasticsearch.Config)
		wantErrCode string
	}{
		{
			name:        "valid configuration",
			modify:      func(c *elasticsearch.Config) {}, // No modification needed
			wantErrCode: "",
		},
		{
			name: "empty addresses",
			modify: func(c *elasticsearch.Config) {
				c.Addresses = nil
			},
			wantErrCode: elasticsearch.ErrCodeEmptyAddresses,
		},
		{
			name: "empty index name",
			modify: func(c *elasticsearch.Config) {
				c.IndexName = ""
			},
			wantErrCode: elasticsearch.ErrCodeEmptyIndexName,
		},
		{
			name: "missing API key",
			modify: func(c *elasticsearch.Config) {
				c.APIKey = ""
			},
			wantErrCode: elasticsearch.ErrCodeMissingAPIKey,
		},
		{
			name: "invalid API key format",
			modify: func(c *elasticsearch.Config) {
				c.APIKey = "invalid_format"
			},
			wantErrCode: elasticsearch.ErrCodeInvalidFormat,
		},
		{
			name: "weak password",
			modify: func(c *elasticsearch.Config) {
				c.Password = "weak"
			},
			wantErrCode: elasticsearch.ErrCodeWeakPassword,
		},
		{
			name: "invalid retry configuration",
			modify: func(c *elasticsearch.Config) {
				c.Retry.InitialWait = -1
			},
			wantErrCode: elasticsearch.ErrCodeInvalidRetry,
		},
		{
			name: "invalid bulk size",
			modify: func(c *elasticsearch.Config) {
				c.BulkSize = 0
			},
			wantErrCode: elasticsearch.ErrCodeInvalidBulkSize,
		},
		{
			name: "invalid flush interval",
			modify: func(c *elasticsearch.Config) {
				c.FlushInterval = 0
			},
			wantErrCode: elasticsearch.ErrCodeInvalidFlush,
		},
		{
			name: "invalid TLS configuration",
			modify: func(c *elasticsearch.Config) {
				c.TLS = &elasticsearch.TLSConfig{
					CertFile: "cert.pem",
					KeyFile:  "", // Missing key file
				}
			},
			wantErrCode: elasticsearch.ErrCodeInvalidTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := testConfig()
			tt.modify(cfg)

			err := cfg.Validate()
			if tt.wantErrCode == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			var configErr *elasticsearch.ConfigError
			require.ErrorAs(t, err, &configErr)
			require.Equal(t, tt.wantErrCode, configErr.Code)
		})
	}
}

func TestNewConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     []elasticsearch.Option
		expected *elasticsearch.Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &elasticsearch.Config{
				Addresses: []string{elasticsearch.DefaultAddresses},
				IndexName: elasticsearch.DefaultIndexName,
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     elasticsearch.DefaultRetryEnabled,
					InitialWait: elasticsearch.DefaultInitialWait,
					MaxWait:     elasticsearch.DefaultMaxWait,
					MaxRetries:  elasticsearch.DefaultMaxRetries,
				},
				BulkSize:      elasticsearch.DefaultBulkSize,
				FlushInterval: elasticsearch.DefaultFlushInterval,
			},
		},
		{
			name: "with custom addresses",
			opts: []elasticsearch.Option{
				elasticsearch.WithAddresses([]string{"http://custom:9200"}),
			},
			expected: &elasticsearch.Config{
				Addresses: []string{"http://custom:9200"},
				IndexName: elasticsearch.DefaultIndexName,
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     elasticsearch.DefaultRetryEnabled,
					InitialWait: elasticsearch.DefaultInitialWait,
					MaxWait:     elasticsearch.DefaultMaxWait,
					MaxRetries:  elasticsearch.DefaultMaxRetries,
				},
				BulkSize:      elasticsearch.DefaultBulkSize,
				FlushInterval: elasticsearch.DefaultFlushInterval,
			},
		},
		{
			name: "with custom API key",
			opts: []elasticsearch.Option{
				elasticsearch.WithAPIKey("custom_id:custom_key"),
			},
			expected: &elasticsearch.Config{
				Addresses: []string{elasticsearch.DefaultAddresses},
				APIKey:    "custom_id:custom_key",
				IndexName: elasticsearch.DefaultIndexName,
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     elasticsearch.DefaultRetryEnabled,
					InitialWait: elasticsearch.DefaultInitialWait,
					MaxWait:     elasticsearch.DefaultMaxWait,
					MaxRetries:  elasticsearch.DefaultMaxRetries,
				},
				BulkSize:      elasticsearch.DefaultBulkSize,
				FlushInterval: elasticsearch.DefaultFlushInterval,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := elasticsearch.NewConfig()
			for _, opt := range tt.opts {
				opt(cfg)
			}
			require.Equal(t, tt.expected, cfg)
		})
	}
}
