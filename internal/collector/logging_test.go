package collector_test

import (
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/mock"
)

// TestConfigureLogging tests the ConfigureLogging function
func TestConfigureLogging(t *testing.T) {
	c := colly.NewCollector()
	testLogger := logger.NewMockLogger()

	// Set up mock expectations for all logging events
	testLogger.On("Debug", "Requesting URL", "url", "http://invalid.example.com").Return()
	testLogger.On("Error", "Error occurred", "url", "http://invalid.example.com", "error", mock.Anything).Return()

	// Call the ConfigureLogging function
	collector.ConfigureLogging(c, testLogger)

	// Create a channel to signal when the error occurs
	errorChan := make(chan bool)
	c.OnError(func(_ *colly.Response, _ error) {
		errorChan <- true
	})

	// Trigger a request to see if logging works
	go c.Visit("http://invalid.example.com")

	// Wait for the error to occur or timeout
	select {
	case <-errorChan:
		// Error occurred as expected
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for error")
	}

	// Give a small delay to ensure all logging is complete
	time.Sleep(100 * time.Millisecond)

	// Verify that the expected log messages were received
	testLogger.AssertExpectations(t)
}
