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
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing addresses",
			config: &Config{
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing index name",
			config: &Config{
				Addresses:     "http://localhost:9200",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid API key format",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "invalid",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid retry configuration",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   0,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid bulk size",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      0,
				FlushInterval: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid flush interval",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without certificate",
			config: &Config{
				Addresses:     "http://localhost:9200",
				IndexName:     "test",
				APIKey:        "test_id:test_key",
				RetryEnabled:  true,
				InitialWait:   1 * time.Second,
				MaxWait:       5 * time.Second,
				MaxRetries:    3,
				BulkSize:      1000,
				FlushInterval: 30 * time.Second,
				TLSEnabled:    true,
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
				Addresses:     DefaultAddresses,
				IndexName:     DefaultIndexName,
				RetryEnabled:  DefaultRetryEnabled,
				InitialWait:   DefaultInitialWait,
				MaxWait:       DefaultMaxWait,
				MaxRetries:    DefaultMaxRetries,
				BulkSize:      DefaultBulkSize,
				FlushInterval: DefaultFlushInterval,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithAddresses("http://custom:9200"),
				WithIndexName("custom"),
				WithAPIKey("custom_id:custom_key"),
				WithRetryEnabled(false),
				WithInitialWait(2 * time.Second),
				WithMaxWait(10 * time.Second),
				WithMaxRetries(5),
				WithBulkSize(2000),
				WithFlushInterval(60 * time.Second),
				WithTLSEnabled(true),
				WithTLSCertFile("cert.pem"),
				WithTLSKeyFile("key.pem"),
				WithTLSCAFile("ca.pem"),
				WithTLSInsecureSkipVerify(true),
			},
			expected: &Config{
				Addresses:             "http://custom:9200",
				IndexName:             "custom",
				APIKey:                "custom_id:custom_key",
				RetryEnabled:          false,
				InitialWait:           2 * time.Second,
				MaxWait:               10 * time.Second,
				MaxRetries:            5,
				BulkSize:              2000,
				FlushInterval:         60 * time.Second,
				TLSEnabled:            true,
				TLSCertFile:           "cert.pem",
				TLSKeyFile:            "key.pem",
				TLSCAFile:             "ca.pem",
				TLSInsecureSkipVerify: true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
