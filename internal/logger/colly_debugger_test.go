package logger_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestCollyDebugger_Event(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	event := &debug.Event{
		Type:        "test",
		RequestID:   1,
		CollectorID: 2,
	}

	// Set up expectations
	mockLogger.EXPECT().Debug("Colly event",
		"type", event.Type,
		"requestID", event.RequestID,
		"collectorID", event.CollectorID,
	).Times(1)

	// Test event handling
	debugger.Event(event)
}

func TestCollyDebugger_OnRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	testURL, _ := url.Parse("http://example.com")
	req := &colly.Request{
		URL:     testURL,
		Method:  "GET",
		Headers: &http.Header{},
	}

	// Set up expectations
	mockLogger.EXPECT().Debug("Request",
		"url", "http://example.com",
		"method", "GET",
		"headers", &http.Header{},
	).Times(1)

	// Test request handling
	debugger.OnRequest(req)
}

func TestCollyDebugger_OnResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	testURL, _ := url.Parse("http://example.com")
	req := &colly.Request{URL: testURL}
	resp := &colly.Response{
		Request:    req,
		StatusCode: 200,
		Headers:    &http.Header{},
	}

	// Set up expectations
	mockLogger.EXPECT().Info("Response",
		"url", "http://example.com",
		"status", 200,
		"headers", &http.Header{},
	).Times(1)

	// Test response handling
	debugger.OnResponse(resp)
}

func TestCollyDebugger_OnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
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
	mockLogger.EXPECT().Error("Error",
		"url", "http://example.com",
		"status", 404,
		"error", "test error",
	).Times(1)

	// Test error handling
	debugger.OnError(resp, testErr)
}

func TestCollyDebugger_OnEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	debugger := &logger.CollyDebugger{Logger: mockLogger}

	event := &debug.Event{
		Type:        "test",
		RequestID:   1,
		CollectorID: 2,
	}

	// Set up expectations
	mockLogger.EXPECT().Info("Event",
		"type", event.Type,
		"requestID", event.RequestID,
		"collectorID", event.CollectorID,
	).Times(1)

	// Test event handling
	debugger.OnEvent(event)
}
