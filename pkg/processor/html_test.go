package processor_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/pkg/logger"
	"github.com/jonesrussell/gocrawl/pkg/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger() logger.Interface {
	return logger.NewNoOp()
}

func newTestMetricsCollector() processor.MetricsCollector {
	return processor.NewMetricsCollector()
}

func TestHTMLProcessor_Process(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantErr  bool
		validate func(t *testing.T, content *processor.Content)
	}{
		{
			name: "valid HTML with all fields",
			html: `
				<!DOCTYPE html>
				<html>
				<head>
					<link rel="canonical" href="http://example.com/article">
					<meta property="og:url" content="http://example.com/article">
					<meta name="author" content="John Doe">
					<meta name="description" content="Test article">
				</head>
				<body>
					<h1>Test Title</h1>
					<article>Test body content</article>
					<div class="author">John Doe</div>
					<time>2024-03-30T12:00:00Z</time>
					<div class="categories">Tech News</div>
					<div class="tags">golang testing</div>
				</body>
				</html>
			`,
			wantErr: false,
			validate: func(t *testing.T, content *processor.Content) {
				assert.Equal(t, "Test Title", content.Title)
				assert.Equal(t, "Test body content", content.Body)
				assert.Equal(t, "http://example.com/article", content.URL)
				assert.Equal(t, "John Doe", content.Author)
				assert.Equal(t, []string{"Tech", "News"}, content.Categories)
				assert.Equal(t, []string{"golang", "testing"}, content.Tags)
				assert.NotZero(t, content.PublishedAt)
				assert.Contains(t, content.Metadata, "author")
				assert.Contains(t, content.Metadata, "description")
			},
		},
		{
			name:    "invalid HTML",
			html:    "not html",
			wantErr: true,
		},
		{
			name: "missing required title",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<article>Test body content</article>
				</body>
				</html>
			`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(map[string]string{
				"title":        "h1",
				"body":         "article",
				"author":       ".author",
				"published_at": "time",
				"categories":   ".categories",
				"tags":         ".tags",
			}, newTestLogger(), newTestMetricsCollector())

			processed, err := p.Process([]byte(tt.html))
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, processed)
			if tt.validate != nil {
				tt.validate(t, &processed.Content)
			}
		})
	}
}

func TestHTMLProcessor_ExtractTime(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		want     time.Time
	}{
		{
			name: "RFC3339 format",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Test Title</h1>
					<time>2024-03-30T12:00:00Z</time>
				</body>
				</html>
			`,
			selector: "time",
			want:     time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "RFC1123 format",
			html:     `<time>Sun, 30 Mar 2024 12:00:00 UTC</time>`,
			selector: "time",
			want:     time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "custom format",
			html:     `<h1>Test Title</h1><time>2024-03-30 12:00:00</time>`,
			selector: "time",
			want:     time.Date(2024, 3, 30, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "empty selector",
			html:     `<time>2024-03-30T12:00:00Z</time>`,
			selector: "",
			want:     time.Time{},
		},
		{
			name:     "empty content",
			html:     `<time></time>`,
			selector: "time",
			want:     time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(map[string]string{
				"published_at": tt.selector,
			}, newTestLogger(), newTestMetricsCollector())

			processed, err := p.Process([]byte(tt.html))
			require.NoError(t, err)
			assert.Equal(t, tt.want, processed.Content.PublishedAt)
		})
	}
}

func TestHTMLProcessor_ExtractList(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		selector string
		want     []string
	}{
		{
			name:     "single item",
			html:     `<div class="categories">Tech</div>`,
			selector: ".categories",
			want:     []string{"Tech"},
		},
		{
			name:     "multiple items",
			html:     `<div class="tags">golang testing</div>`,
			selector: ".tags",
			want:     []string{"golang", "testing"},
		},
		{
			name:     "empty selector",
			html:     `<div class="categories">Tech</div>`,
			selector: "",
			want:     nil,
		},
		{
			name:     "empty content",
			html:     `<div class="categories"></div>`,
			selector: ".categories",
			want:     nil,
		},
		{
			name: "multiple elements",
			html: `
				<div class="categories">Tech</div>
				<div class="categories">News</div>
			`,
			selector: ".categories",
			want:     []string{"Tech", "News"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(map[string]string{
				"categories": tt.selector,
			}, newTestLogger(), newTestMetricsCollector())

			processed, err := p.Process([]byte(tt.html))
			require.NoError(t, err)
			assert.Equal(t, tt.want, processed.Content.Categories)
		})
	}
}

func TestHTMLProcessor_ExtractMetadata(t *testing.T) {
	tests := []struct {
		name string
		html string
		want map[string]string
	}{
		{
			name: "meta tags with name",
			html: `
				<meta name="author" content="John Doe">
				<meta name="description" content="Test article">
			`,
			want: map[string]string{
				"author":      "John Doe",
				"description": "Test article",
			},
		},
		{
			name: "meta tags with property",
			html: `
				<meta property="og:title" content="Test Title">
				<meta property="og:url" content="http://example.com">
			`,
			want: map[string]string{
				"og:title": "Test Title",
				"og:url":   "http://example.com",
			},
		},
		{
			name: "mixed meta tags",
			html: `
				<meta name="author" content="John Doe">
				<meta property="og:title" content="Test Title">
			`,
			want: map[string]string{
				"author":   "John Doe",
				"og:title": "Test Title",
			},
		},
		{
			name: "no meta tags",
			html: `<div>No meta tags</div>`,
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(nil, newTestLogger(), newTestMetricsCollector())

			processed, err := p.Process([]byte(tt.html))
			require.NoError(t, err)
			assert.Equal(t, tt.want, processed.Content.Metadata)
		})
	}
}
