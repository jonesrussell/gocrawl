package collector_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/mock"
)

func TestConfigureLogging(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	c := colly.NewCollector()

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}))
	defer ts.Close()

	// Test request logging
	mockLogger.On("Debug", "Requesting URL",
		"url", ts.URL,
	).Return()

	// Test response logging
	mockLogger.On("Debug", "Received response",
		"url", ts.URL,
		"status", http.StatusOK,
	).Return()

	// Test error logging for localhost:1
	mockLogger.On("Debug", "Requesting URL",
		"url", "http://localhost:1",
	).Return()

	mockLogger.On("Error", "Error occurred",
		"url", "http://localhost:1",
		"error", mock.Anything,
	).Return()

	// Configure logging
	collector.ConfigureLogging(c, mockLogger)

	// Visit test server to trigger request and response callbacks
	err := c.Visit(ts.URL)
	if err != nil {
		return
	}

	// Visit with error to trigger error callback
	errVisit := c.Visit("http://localhost:1")
	if errVisit != nil {
		return
	}
}
