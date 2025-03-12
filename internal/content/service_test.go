package content_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestExtractContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	// Create a test HTML element
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html>
			<head>
				<title>Test Title</title>
				<meta property="og:type" content="article">
				<script type="application/ld+json">
					{
						"@type": "Article",
						"dateCreated": "2024-03-03T12:00:00Z",
						"name": "Test Article"
					}
				</script>
			</head>
			<body>
				<h1>Test Heading</h1>
				<div class="content">Test Content</div>
			</body>
		</html>
	`))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	e := &colly.HTMLElement{
		Request: &colly.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		},
		DOM: doc.Selection,
	}

	// Set up expectations
	mockLogger.EXPECT().Debug("Extracting content", "url", "http://example.com/test").Times(1)
	mockLogger.EXPECT().Debug("Successfully parsed date",
		"source", "2024-03-03T12:00:00Z",
		"format", "2006-01-02T15:04:05Z",
		"result", gomock.Any(),
	).Times(1)
	mockLogger.EXPECT().Debug("Extracted content",
		"id", gomock.Any(),
		"title", "Test Article",
		"url", "http://example.com/test",
		"type", "Article",
		"created_at", gomock.Any(),
	).Times(1)

	// Test content extraction
	content := service.ExtractContent(e)
	assert.NotNil(t, content)
	assert.Equal(t, "Test Article", content.Title)
	assert.Equal(t, "Article", content.Type)
	assert.Contains(t, content.Body, "Test Content")
}

func TestExtractMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	// Create a test HTML element
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html>
			<head>
				<meta property="og:type" content="article">
				<meta property="og:title" content="Test Title">
				<meta name="description" content="Test Description">
			</head>
		</html>
	`))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	e := &colly.HTMLElement{
		DOM: doc.Selection,
	}

	// Test metadata extraction
	metadata := service.ExtractMetadata(e)
	assert.NotNil(t, metadata)
	assert.Equal(t, "article", metadata["type"])
	assert.Equal(t, "Test Title", metadata["title"])
	assert.Equal(t, "Test Description", metadata["description"])
}

func TestDetermineContentType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	tests := []struct {
		name       string
		url        string
		metadata   map[string]any
		jsonLDType string
		expected   string
	}{
		{
			name:       "uses JSON-LD type",
			url:        "http://example.com/test",
			metadata:   nil,
			jsonLDType: "Article",
			expected:   "Article",
		},
		{
			name:       "uses metadata type",
			url:        "http://example.com/test",
			metadata:   map[string]any{"type": "BlogPost"},
			jsonLDType: "",
			expected:   "BlogPost",
		},
		{
			name:       "detects category from URL",
			url:        "http://example.com/category/test",
			metadata:   nil,
			jsonLDType: "",
			expected:   "category",
		},
		{
			name:       "defaults to webpage",
			url:        "http://example.com/test",
			metadata:   nil,
			jsonLDType: "",
			expected:   "webpage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType := service.DetermineContentType(tt.url, tt.metadata, tt.jsonLDType)
			assert.Equal(t, tt.expected, contentType)
		})
	}
}

func TestService_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes HTML tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name:     "trims whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.Process(t.Context(), tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_ProcessBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "processes multiple strings",
			input: []string{
				"<p>Hello <b>World</b></p>",
				"  Goodbye    World  ",
			},
			expected: []string{
				"Hello World",
				"Goodbye World",
			},
		},
		{
			name:     "handles empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "handles empty strings",
			input: []string{
				"",
				"  ",
			},
			expected: []string{
				"",
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ProcessBatch(t.Context(), tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_ProcessWithMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	service := content.NewService(mockLogger)

	tests := []struct {
		name     string
		input    string
		metadata map[string]string
		expected string
	}{
		{
			name:  "processes with metadata",
			input: "<p>Hello <b>World</b></p>",
			metadata: map[string]string{
				"source": "test",
				"type":   "article",
			},
			expected: "Hello World",
		},
		{
			name:     "processes without metadata",
			input:    "<p>Hello <b>World</b></p>",
			metadata: nil,
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations for logging
			if tt.metadata != nil {
				mockLogger.EXPECT().Debug("Processing content with metadata",
					"source", tt.metadata["source"],
					"type", tt.metadata["type"],
				).Times(1)
			}

			result := service.ProcessWithMetadata(t.Context(), tt.input, tt.metadata)
			assert.Equal(t, tt.expected, result)
		})
	}
}
