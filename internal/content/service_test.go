package content_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExtractContent(t *testing.T) {
	// Create mock logger
	mockLogger := logger.NewMockLogger()
	mockLogger.On("Debug", "Extracting content", "url", "http://example.com").Return()
	mockLogger.On("Debug", "Extracted content",
		"id", mock.AnythingOfType("string"),
		"title", "Test Content",
		"url", "http://example.com",
		"type", "webpage",
		"created_at", mock.AnythingOfType("time.Time")).Return()

	// Create service
	svc := content.NewService(mockLogger)

	// Create test HTML
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Content</title>
			<script type="application/ld+json">
			{
				"@type": "WebPage",
				"name": "Test Content",
				"dateCreated": "2024-03-03T12:00:00Z",
				"description": "Test Description"
			}
			</script>
		</head>
		<body>
			<h1>Test Content</h1>
			<p>Test body content</p>
		</body>
		</html>
	`

	// Create colly HTMLElement
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	e := &colly.HTMLElement{
		DOM:     doc.Selection,
		Request: &colly.Request{URL: &url.URL{Path: "http://example.com"}},
	}

	// Extract content
	result := svc.ExtractContent(e)

	// Verify result
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Test Content", result.Title)
	assert.Equal(t, "http://example.com", result.URL)
	assert.Equal(t, "webpage", result.Type)
	assert.Contains(t, result.Body, "Test body content")
	assert.NotZero(t, result.CreatedAt)

	// Verify mock expectations
	mockLogger.AssertExpectations(t)
}

func TestExtractMetadata(t *testing.T) {
	// Create mock logger
	mockLogger := logger.NewMockLogger()

	// Create service
	svc := content.NewService(mockLogger)

	// Create test HTML with various metadata
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta property="og:title" content="OG Title">
			<meta property="og:description" content="OG Description">
			<meta name="twitter:title" content="Twitter Title">
			<meta name="twitter:description" content="Twitter Description">
			<meta name="author" content="Test Author">
		</head>
		<body>
			<h1>Test Content</h1>
		</body>
		</html>
	`

	// Create colly HTMLElement
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
	e := &colly.HTMLElement{
		DOM:     doc.Selection,
		Request: &colly.Request{URL: &url.URL{Path: "http://example.com"}},
	}

	// Extract metadata
	metadata := svc.ExtractMetadata(e)

	// Verify metadata
	assert.Equal(t, "OG Title", metadata["title"])
	assert.Equal(t, "OG Description", metadata["description"])
	assert.Equal(t, "Twitter Title", metadata["twitter:title"])
	assert.Equal(t, "Twitter Description", metadata["twitter:description"])
	assert.Equal(t, "Test Author", metadata["author"])
}

func TestDetermineContentType(t *testing.T) {
	// Create mock logger
	mockLogger := logger.NewMockLogger()

	// Create service
	svc := content.NewService(mockLogger)

	tests := []struct {
		name     string
		url      string
		metadata map[string]interface{}
		jsonType string
		expected string
	}{
		{
			name:     "Category URL",
			url:      "http://example.com/category/news",
			metadata: map[string]interface{}{},
			jsonType: "",
			expected: "category",
		},
		{
			name:     "Tag URL",
			url:      "http://example.com/tag/tech",
			metadata: map[string]interface{}{},
			jsonType: "",
			expected: "tag",
		},
		{
			name:     "Author URL",
			url:      "http://example.com/author/john",
			metadata: map[string]interface{}{},
			jsonType: "",
			expected: "author",
		},
		{
			name:     "JSON-LD type takes precedence",
			url:      "http://example.com/category/news",
			metadata: map[string]interface{}{},
			jsonType: "Article",
			expected: "Article",
		},
		{
			name: "Metadata type takes precedence over URL",
			url:  "http://example.com/category/news",
			metadata: map[string]interface{}{
				"type": "BlogPost",
			},
			jsonType: "",
			expected: "BlogPost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.DetermineContentType(tt.url, tt.metadata, tt.jsonType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
