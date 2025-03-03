package collector_test

import (
	"testing"

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
	testLogger.On("Debug", "Requesting URL", mock.Anything, mock.Anything).Return()
	testLogger.On("Debug", "Received response", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	testLogger.On("Error", "Error occurred", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	// Call the ConfigureLogging function
	collector.ConfigureLogging(c, testLogger)

	// Trigger a request to see if logging works
	c.Visit("http://example.com")

	// Verify that the expected log messages were received
	testLogger.AssertCalled(t, "Debug", "Requesting URL", "url", mock.Anything)
}
