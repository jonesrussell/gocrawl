package elasticsearch_test

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/elasticsearch"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	validAPIKey := base64.StdEncoding.EncodeToString([]byte("test_key"))

	tests := []struct {
		name    string
		config  *elasticsearch.Config
		wantErr bool
		errCode string
	}{
		{
			name: "valid configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    validAPIKey,
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
				APIKey:    validAPIKey,
				IndexName: "test",
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeEmptyAddresses,
		},
		{
			name: "empty index name",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    validAPIKey,
				IndexName: "",
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeEmptyIndexName,
		},
		{
			name: "empty API key",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "",
				IndexName: "test",
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeMissingAPIKey,
		},
		{
			name: "invalid API key format",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "not_base64_encoded",
				IndexName: "test",
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeInvalidFormat,
		},
		{
			name: "weak password",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
				IndexName: "test",
				Password:  "weak",
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeWeakPassword,
		},
		{
			name: "invalid retry configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    validAPIKey,
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
			errCode: elasticsearch.ErrCodeInvalidRetry,
		},
		{
			name: "invalid bulk size",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    validAPIKey,
				IndexName: "test",
				BulkSize:  0,
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeInvalidBulkSize,
		},
		{
			name: "invalid flush interval",
			config: &elasticsearch.Config{
				Addresses:     []string{"http://localhost:9200"},
				APIKey:        validAPIKey,
				IndexName:     "test",
				FlushInterval: 0,
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeInvalidFlush,
		},
		{
			name: "invalid TLS configuration",
			config: &elasticsearch.Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    validAPIKey,
				IndexName: "test",
				TLS: &elasticsearch.TLSConfig{
					Enabled:  true,
					CertFile: "",
					KeyFile:  "key.pem",
				},
			},
			wantErr: true,
			errCode: elasticsearch.ErrCodeInvalidTLS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				var configErr *elasticsearch.ConfigError
				if errors.As(err, &configErr) {
					require.Equal(t, tt.errCode, configErr.Code)
				} else {
					t.Errorf("expected ConfigError, got %T", err)
				}
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
