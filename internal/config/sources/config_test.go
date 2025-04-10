package sources

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
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			config: &Config{
				BaseURL:      "",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &Config{
				BaseURL:      "invalid-url",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "negative max depth",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     -1,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "zero parallelism",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  0,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    -1 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      0,
				UserAgent:    "test-agent",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "empty user agent",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "",
				AllowDomains: "example.com",
			},
			wantErr: true,
		},
		{
			name: "empty allow domains",
			config: &Config{
				BaseURL:      "http://example.com",
				MaxDepth:     2,
				Parallelism:  2,
				RateLimit:    2 * time.Second,
				Timeout:      30 * time.Second,
				UserAgent:    "test-agent",
				AllowDomains: "",
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
				MaxDepth:     DefaultMaxDepth,
				Parallelism:  DefaultParallelism,
				RateLimit:    DefaultRateLimit,
				Timeout:      DefaultTimeout,
				UserAgent:    DefaultUserAgent,
				AllowDomains: DefaultAllowDomains,
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithBaseURL("http://example.com"),
				WithMaxDepth(3),
				WithParallelism(4),
				WithRateLimit(5 * time.Second),
				WithTimeout(60 * time.Second),
				WithUserAgent("custom-agent"),
				WithAllowDomains("example.com,test.com"),
				WithDisallowedURLFilters([]string{"*.pdf", "*.jpg"}),
			},
			expected: &Config{
				BaseURL:              "http://example.com",
				MaxDepth:             3,
				Parallelism:          4,
				RateLimit:            5 * time.Second,
				Timeout:              60 * time.Second,
				UserAgent:            "custom-agent",
				AllowDomains:         "example.com,test.com",
				DisallowedURLFilters: []string{"*.pdf", "*.jpg"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
