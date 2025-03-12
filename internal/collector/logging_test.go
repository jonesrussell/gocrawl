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
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	// Test request logging
	mockLogger.EXPECT().Debug("Requesting URL",
		"url", ts.URL,
	).Times(1)

	// Test response logging
	mockLogger.EXPECT().Debug("Received response",
		"url", ts.URL,
		"status", 200,
	).Times(1)

	// Test error logging
	mockLogger.EXPECT().Error("Error occurred",
		"url", "http://localhost:1",
		"error", "Get \"http://localhost:1\": dial tcp [::1]:1: connect: connection refused",
	).Times(1)

	// Configure logging
	collector.ConfigureLogging(c, mockLogger)

	// Visit test server to trigger request and response callbacks
	c.Visit(ts.URL)

	// Visit with error to trigger error callback
	c.Visit("http://localhost:1")
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
