package elasticsearch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
			name: "missing addresses",
			config: &Config{
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
			wantErr: true,
		},
		{
			name: "missing index name",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				APIKey:    "test_id:test_key",
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
			wantErr: true,
		},
		{
			name: "missing API key",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
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
			wantErr: true,
		},
		{
			name: "invalid API key format",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "invalid",
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
			wantErr: true,
		},
		{
			name: "invalid retry configuration",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid bulk size",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
				BulkSize:      0,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid flush interval",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
				FlushInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without certificate",
			config: &Config{
				Addresses: []string{"http://localhost:9200"},
				IndexName: "test",
				APIKey:    "test_id:test_key",
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
				TLS: &TLSConfig{
					Enabled: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
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

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     []Option
		expected *Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &Config{
				Addresses: []string{DefaultAddresses},
				IndexName: DefaultIndexName,
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     DefaultRetryEnabled,
					InitialWait: DefaultInitialWait,
					MaxWait:     DefaultMaxWait,
					MaxRetries:  DefaultMaxRetries,
				},
				BulkSize:      DefaultBulkSize,
				FlushInterval: DefaultFlushInterval,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithAddresses([]string{"http://custom:9200"}),
				WithIndexName("custom"),
				WithAPIKey("custom_id:custom_key"),
				WithRetryEnabled(false),
				WithInitialWait(2 * time.Second),
				WithMaxWait(10 * time.Second),
				WithMaxRetries(5),
				WithBulkSize(2000),
				WithFlushInterval(60 * time.Second),
			},
			expected: &Config{
				Addresses: []string{"http://custom:9200"},
				IndexName: "custom",
				APIKey:    "custom_id:custom_key",
				Retry: struct {
					Enabled     bool          `yaml:"enabled"`
					InitialWait time.Duration `yaml:"initial_wait"`
					MaxWait     time.Duration `yaml:"max_wait"`
					MaxRetries  int           `yaml:"max_retries"`
				}{
					Enabled:     false,
					InitialWait: 2 * time.Second,
					MaxWait:     10 * time.Second,
					MaxRetries:  5,
				},
				BulkSize:      2000,
				FlushInterval: 60 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewConfig()
			for _, opt := range tt.opts {
				opt(cfg)
			}
			require.Equal(t, tt.expected, cfg)
		})
	}
}

func TestNewConfig(t *testing.T) {
	expectedAddresses := []string{"http://localhost:9200"}
	cfg := NewConfig()

	require.NotNil(t, cfg)
	require.Equal(t, expectedAddresses, cfg.Addresses)
	require.Empty(t, cfg.APIKey)
	require.Empty(t, cfg.Username)
	require.Empty(t, cfg.Password)
	require.Empty(t, cfg.IndexName)
	require.Empty(t, cfg.Cloud.ID)
	require.Empty(t, cfg.Cloud.APIKey)
	require.NotNil(t, cfg.TLS)
	require.False(t, cfg.TLS.Enabled)
	require.NotNil(t, cfg.Retry)
	require.True(t, cfg.Retry.Enabled)
	require.Equal(t, time.Second, cfg.Retry.InitialWait)
	require.Equal(t, time.Minute, cfg.Retry.MaxWait)
	require.Equal(t, 3, cfg.Retry.MaxRetries)
	require.Equal(t, 1000, cfg.BulkSize)
	require.Equal(t, 30*time.Second, cfg.FlushInterval)
}
