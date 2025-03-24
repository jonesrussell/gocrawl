package collector_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestConfigureLogging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
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
	mockLogger.EXPECT().Debug("Requesting URL",
		"url", ts.URL,
	).Times(1)

	// Test response logging
	mockLogger.EXPECT().Debug("Received response",
		"url", ts.URL,
		"status", http.StatusOK,
	).Times(1)

	// Test error logging for localhost:1
	mockLogger.EXPECT().Debug("Requesting URL",
		"url", "http://localhost:1",
	).Times(1)

	mockLogger.EXPECT().Error("Error occurred",
		"url", "http://localhost:1",
		"error", gomock.Any(),
	).Times(1)

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
