package elasticsearch_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *elasticsearch.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "test",
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
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty addresses",
			config: &elasticsearch.Config{
				Addresses: []string{},
				APIKey:    "test_id:test_key",
				IndexName: "test",
			},
			wantErr: true,
		},
		{
			name: "empty index name",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "",
			},
			wantErr: true,
		},
		{
			name: "empty API key",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "",
				IndexName: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid API key format",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "invalid",
				IndexName: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid retry configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "test",
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     true,
					InitialWait: 0,
					MaxWait:     5 * time.Second,
					MaxRetries:  3,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid bulk size",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "test",
				BulkSize:  0,
			},
			wantErr: true,
		},
		{
			name: "invalid flush interval",
			config: &elasticsearch.Config{
				Addresses:     []string{"http://localhost:9200"},
				APIKey:        "test_id:test_key",
				IndexName:     "test",
				FlushInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid TLS configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "test",
				TLS: &elasticsearch.TLSConfig{
					Enabled:  true,
					CertFile: "",
					KeyFile:  "key.pem",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
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
