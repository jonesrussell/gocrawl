package crawler

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
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "invalid max depth",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    0,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   0,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "invalid parallelism",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 0,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     0,
				MaxBodySize: 10 * 1024 * 1024,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "invalid max body size",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 0,
				SourceFile:  "sources.yml",
			},
			wantErr: true,
		},
		{
			name: "missing source file",
			config: &Config{
				BaseURL:     "http://example.com",
				MaxDepth:    2,
				RateLimit:   2 * time.Second,
				Parallelism: 2,
				Timeout:     30 * time.Second,
				MaxBodySize: 10 * 1024 * 1024,
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
				MaxDepth:          DefaultMaxDepth,
				RateLimit:         DefaultRateLimit,
				Parallelism:       DefaultParallelism,
				UserAgent:         DefaultUserAgent,
				Timeout:           DefaultTimeout,
				MaxBodySize:       DefaultMaxBodySize,
				AllowedDomains:    []string{},
				DisallowedDomains: []string{},
			},
		},
		{
			name: "custom configuration",
			opts: []Option{
				WithBaseURL("http://example.com"),
				WithMaxDepth(5),
				WithRateLimit(5 * time.Second),
				WithParallelism(5),
				WithUserAgent("custom/1.0"),
				WithTimeout(60 * time.Second),
				WithMaxBodySize(20 * 1024 * 1024),
				WithAllowedDomains([]string{"example.com"}),
				WithDisallowedDomains([]string{"forbidden.com"}),
				WithSourceFile("custom.yml"),
			},
			expected: &Config{
				BaseURL:           "http://example.com",
				MaxDepth:          5,
				RateLimit:         5 * time.Second,
				Parallelism:       5,
				UserAgent:         "custom/1.0",
				Timeout:           60 * time.Second,
				MaxBodySize:       20 * 1024 * 1024,
				AllowedDomains:    []string{"example.com"},
				DisallowedDomains: []string{"forbidden.com"},
				SourceFile:        "custom.yml",
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
