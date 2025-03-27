package logger_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/testutils"
)

func TestCollyDebugger_Event(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	event := &debug.Event{
		Type:        "test",
		RequestID:   1,
		CollectorID: 2,
	}

	// Set up expectations
	mockLogger.On("Debug", "Colly event",
		"type", event.Type,
		"requestID", event.RequestID,
		"collectorID", event.CollectorID,
	).Return()

	// Test event handling
	debugger.Event(event)
}

func TestCollyDebugger_OnRequest(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	testURL, _ := url.Parse("http://example.com")
	req := &colly.Request{
		URL:     testURL,
		Method:  "GET",
		Headers: &http.Header{},
	}

	// Set up expectations
	mockLogger.On("Debug", "Request",
		"url", "http://example.com",
		"method", "GET",
		"headers", &http.Header{},
	).Return()

	// Test request handling
	debugger.OnRequest(req)
}

func TestCollyDebugger_OnResponse(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	testURL, _ := url.Parse("http://example.com")
	req := &colly.Request{URL: testURL}
	resp := &colly.Response{
		Request:    req,
		StatusCode: 200,
		Headers:    &http.Header{},
	}

	// Set up expectations
	mockLogger.On("Info", "Response",
		"url", "http://example.com",
		"status", 200,
		"headers", &http.Header{},
	).Return()

	// Test response handling
	debugger.OnResponse(resp)
}

func TestCollyDebugger_OnError(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	testURL, _ := url.Parse("http://example.com")
	req := &colly.Request{URL: testURL}
	resp := &colly.Response{
		Request:    req,
		StatusCode: 404,
		Headers:    &http.Header{},
	}
	testErr := errors.New("test error")

	// Set up expectations
	mockLogger.On("Error", "Error",
		"url", "http://example.com",
		"status", 404,
		"error", "test error",
	).Return()

	// Test error handling
	debugger.OnError(resp, testErr)
}

func TestCollyDebugger_OnEvent(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	event := &debug.Event{
		Type:        "test",
		RequestID:   1,
		CollectorID: 2,
	}

	// Set up expectations
	mockLogger.On("Info", "Event",
		"type", event.Type,
		"requestID", event.RequestID,
		"collectorID", event.CollectorID,
	).Return()

	// Test event handling
	debugger.OnEvent(event)
}
