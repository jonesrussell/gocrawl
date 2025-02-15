package logger_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestMockCollyDebugger(t *testing.T) {
	debugger := logger.NewMockCollyDebugger()

	t.Run("Init", func(_ *testing.T) {
		err := debugger.Init()
		assert.NoError(t, err)
	})

	t.Run("OnRequest", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{
			URL:     testURL,
			Method:  "GET",
			Headers: &http.Header{},
		}
		debugger.OnRequest(req)
	})

	t.Run("OnResponse", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 200,
			Headers:    &http.Header{},
		}
		debugger.OnResponse(resp)
	})

	t.Run("OnError", func(_ *testing.T) {
		testURL, _ := url.Parse("http://example.com")
		req := &colly.Request{URL: testURL}
		resp := &colly.Response{
			Request:    req,
			StatusCode: 404,
			Headers:    &http.Header{},
		}
		debugger.OnError(resp, errors.New("test error"))
	})

	t.Run("OnEvent", func(_ *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		debugger.OnEvent(event)
	})

	t.Run("Event", func(_ *testing.T) {
		event := &debug.Event{
			Type:        "test",
			RequestID:   1,
			CollectorID: 1,
		}
		debugger.Event(event)
	})
}
