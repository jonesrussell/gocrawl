package logger_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestCollyDebugger(t *testing.T) {
	mockLogger := logger.NewMockCustomLogger()
	debugger := &logger.CollyDebugger{
		Logger: mockLogger,
	}

	t.Run("Init", func(t *testing.T) {
		if err := debugger.Init(); err != nil {
			t.Errorf("Init() error = %v", err)
		}
	})

	t.Run("Event with nil logger", func(t *testing.T) {
		nilDebugger := &logger.CollyDebugger{Logger: nil}
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		nilDebugger.Event(event) // Should not panic
	})

	t.Run("Event with logger", func(t *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		debugger.Event(event)
	})

	t.Run("OnRequest", func(t *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{
			URL:     testURL,
			Method:  "GET",
			Headers: &http.Header{},
		}
		debugger.OnRequest(req)
	})

	t.Run("OnResponse", func(t *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 200,
			Headers:    &http.Header{},
		}
		debugger.OnResponse(resp)
	})

	t.Run("OnError", func(t *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 404,
			Headers:    &http.Header{},
		}
		debugger.OnError(resp, nil)
	})

	t.Run("OnEvent", func(t *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		debugger.OnEvent(event)
	})
}
