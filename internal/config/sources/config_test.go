package sources_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/internal/config/sources"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *sources.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
			config: &sources.Config{
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
		opts     []sources.Option
		expected *sources.Config
	}{
		{
			name: "default configuration",
			opts: nil,
			expected: &sources.Config{
				MaxDepth:     sources.DefaultMaxDepth,
				Parallelism:  sources.DefaultParallelism,
				RateLimit:    sources.DefaultRateLimit,
				Timeout:      sources.DefaultTimeout,
				UserAgent:    sources.DefaultUserAgent,
				AllowDomains: sources.DefaultAllowDomains,
			},
		},
		{
			name: "custom configuration",
			opts: []sources.Option{
				sources.WithBaseURL("http://example.com"),
				sources.WithMaxDepth(3),
				sources.WithParallelism(4),
				sources.WithRateLimit(5 * time.Second),
				sources.WithTimeout(60 * time.Second),
				sources.WithUserAgent("custom-agent"),
				sources.WithAllowDomains("example.com,test.com"),
				sources.WithDisallowedURLFilters([]string{"*.pdf", "*.jpg"}),
			},
			expected: &sources.Config{
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

			cfg := sources.New(tt.opts...)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
