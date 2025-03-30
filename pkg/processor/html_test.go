package processor_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/gocrawl/pkg/processor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLProcessor_Process(t *testing.T) {
	selectors := map[string]string{
		"title":        "h1",
		"body":         "article",
		"author":       ".author",
		"published_at": "time",
		"categories":   ".categories span",
		"tags":         ".tags span",
	}

	tests := []struct {
		name     string
		html     string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, processed *processor.ProcessedContent)
	}{
		{
			name: "valid HTML",
			html: `
				<!DOCTYPE html>
				<html>
				<head>
					<link rel="canonical" href="http://example.com/article">
					<meta property="og:url" content="http://example.com/article">
					<meta name="description" content="Test article">
				</head>
				<body>
					<h1>Test Article</h1>
					<article>This is the article body.</article>
					<div class="author">John Doe</div>
					<time>2024-03-20T10:00:00Z</time>
					<div class="categories">
						<span>Technology</span>
						<span>Programming</span>
					</div>
					<div class="tags">
						<span>go</span>
						<span>testing</span>
					</div>
				</body>
				</html>
			`,
			wantErr: false,
			validate: func(t *testing.T, processed *processor.ProcessedContent) {
				assert.Equal(t, "Test Article", processed.Content.Title)
				assert.Equal(t, "This is the article body.", processed.Content.Body)
				assert.Equal(t, "http://example.com/article", processed.Content.URL)
				assert.Equal(t, "John Doe", processed.Content.Author)
				assert.Equal(t, time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC), processed.Content.PublishedAt)
				assert.Equal(t, []string{"Technology", "Programming"}, processed.Content.Categories)
				assert.Equal(t, []string{"go", "testing"}, processed.Content.Tags)
				assert.Equal(t, "Test article", processed.Content.Metadata["description"])
			},
		},
		{
			name: "missing required fields",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<div class="author">John Doe</div>
				</body>
				</html>
			`,
			wantErr: true,
			errMsg:  "missing required field: title",
		},
		{
			name:    "invalid HTML",
			html:    `not html`,
			wantErr: true,
			errMsg:  "invalid HTML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(selectors)
			processed, err := p.Process([]byte(tt.html))
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, processed)
			if tt.validate != nil {
				tt.validate(t, processed)
			}
		})
	}
}

func TestHTMLProcessor_ExtractTime(t *testing.T) {
	selectors := map[string]string{
		"title":        "h1",
		"published_at": "time",
	}

	tests := []struct {
		name     string
		html     string
		expected time.Time
	}{
		{
			name: "RFC3339",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Test Article</h1>
					<time>2024-03-20T10:00:00Z</time>
				</body>
				</html>
			`,
			expected: time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "RFC1123",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Test Article</h1>
					<time>Wed, 20 Mar 2024 10:00:00 UTC</time>
				</body>
				</html>
			`,
			expected: time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "custom format",
			html: `
				<!DOCTYPE html>
				<html>
				<body>
					<h1>Test Article</h1>
					<time>2024-03-20 10:00:00</time>
				</body>
				</html>
			`,
			expected: time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.NewHTMLProcessor(selectors)
			processed, err := p.Process([]byte(tt.html))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, processed.Content.PublishedAt)
		})
	}
}
