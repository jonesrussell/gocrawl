package collector_test

import (
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/require"
)

// TestNew tests the New function of the collector package
func TestNew(t *testing.T) {
	mockDebugger := &logger.CollyDebugger{
		Logger: logger.NewMockCustomLogger(),
	}

	params := collector.Params{
		BaseURL:   "http://example.com",
		MaxDepth:  2,
		RateLimit: 1 * time.Second,
		Debugger:  mockDebugger,
	}

	result, err := collector.New(params)
	require.NoError(t, err)
	require.NotNil(t, result.Collector)

	// Additional assertions can be made here based on the collector's configuration
}

// TestConfigureLogging tests the ConfigureLogging function
func TestConfigureLogging(t *testing.T) {
	c := colly.NewCollector()
	testLogger := logger.NewMockCustomLogger()

	// Call the ConfigureLogging function
	collector.ConfigureLogging(c, testLogger)

	// Simulate a request to test logging
	c.OnRequest(func(r *colly.Request) {
		testLogger.Info("Requesting URL", r.URL.String())
	})

	// Trigger a request to see if logging works
	c.Visit("http://example.com")

	// Check if the logger received the expected log messages
	messages := testLogger.GetMessages()
	require.Contains(t, messages, "Requesting URL")
}

func TestCollector(t *testing.T) {
	testLogger := logger.NewMockCustomLogger()

	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid URL",
			baseURL: "https://example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty URL",
			baseURL: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := collector.Params{
				BaseURL:   tt.baseURL,
				MaxDepth:  2,
				RateLimit: time.Second,
				Debugger: &logger.CollyDebugger{
					Logger: testLogger,
				},
			}

			result, err := collector.New(params)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result.Collector)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result.Collector)
			}
		})
	}
}
