package logger_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestCollyDebugger(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	debugger := &logger.CollyDebugger{
		Logger: mockLogger,
	}

	t.Run("Init", func(_ *testing.T) {
		if err := debugger.Init(); err != nil {
			t.Errorf("Init() error = %v", err)
		}
	})

	t.Run("Event with nil logger", func(_ *testing.T) {
		nilDebugger := &logger.CollyDebugger{Logger: nil}
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		nilDebugger.Event(event) // Should not panic
	})

	t.Run("Event with logger", func(_ *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		// Set expectation for the Debug method with the correct format
		mockLogger.On("Debug", "Colly event",
			"type", "test",
			"requestID", uint32(1),
			"collectorID", uint32(1),
		).Return()

		debugger.Event(event)

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})

	t.Run("OnRequest", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{
			URL:     testURL,
			Method:  "GET",
			Headers: &http.Header{},
		}
		// Set expectation for the Debug method in OnRequest
		mockLogger.On("Debug", "Request",
			"url", "http://example.com",
			"method", "GET",
			"headers", &http.Header{},
		).Return()

		debugger.OnRequest(req)

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})

	t.Run("OnResponse", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 200,
			Headers:    &http.Header{},
		}
		// Set expectation for the Info method in OnResponse
		mockLogger.On("Info", "Response",
			"url", "http://example.com",
			"status", 200,
			"headers", &http.Header{},
		).Return()

		debugger.OnResponse(resp)

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})

	t.Run("OnError", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 404,
			Headers:    &http.Header{},
		}
		// Set expectation for the Error method in OnError
		mockLogger.On("Error", "Error",
			"url", "http://example.com",
			"status", 404,
			"error", "unknown error",
		).Return()

		debugger.OnError(resp, errors.New("unknown error"))

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})

	t.Run("OnEvent", func(_ *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		// Set expectation for the Info method in OnEvent
		mockLogger.On("Info", "Event",
			"type", "test",
			"requestID", uint32(1),
			"collectorID", uint32(1),
		).Return()

		debugger.OnEvent(event)

		// Assert that the expectations were met
		mockLogger.AssertExpectations(t)
	})
}
