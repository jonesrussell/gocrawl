package content_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/logger/mock_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mock_logger.NewMockInterface(ctrl)
	svc := content.NewService(mockLogger)

	// Create a test HTML document
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
	<script type="application/ld+json">
	{
		"@type": "Article",
		"name": "Test Article",
		"dateCreated": "2024-03-15T10:00:00Z"
	}
	</script>
	<meta property="article:published_time" content="2024-03-15T10:00:00Z" />
</head>
<body>
	<h1>Test Article</h1>
	<p>Test content</p>
</body>
</html>`

	// Create a mock colly HTMLElement
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	testURL := "http://example.com/test"
	parsedURL, err := url.Parse(testURL)
	require.NoError(t, err)

	req := &colly.Request{
		URL: parsedURL,
		Ctx: colly.NewContext(),
	}
	resp := &colly.Response{
		Request: req,
		Ctx:     req.Ctx,
	}
	e := &colly.HTMLElement{
		Request:  req,
		Response: resp,
		DOM:      doc.Selection,
	}

	// Set up logger expectations in order
	mockLogger.EXPECT().Debug("Extracting content", "url", testURL)
	mockLogger.EXPECT().Debug("Trying to parse date", "value", "2024-03-15T10:00:00Z")
	mockLogger.EXPECT().Debug("Successfully parsed date",
		"source", "2024-03-15T10:00:00Z",
		"format", time.RFC3339,
		"result", "2024-03-15 10:00:00 +0000 UTC",
	)
	mockLogger.EXPECT().Debug("Extracted content",
		"id", gomock.Any(),
		"title", "Test Article",
		"url", testURL,
		"type", "article",
		"created_at", gomock.Any(),
	)

	content := svc.ExtractContent(e)
	require.NotNil(t, content)
	assert.Equal(t, "Test Article", content.Title)
	assert.Equal(t, testURL, content.URL)
	assert.Equal(t, "article", content.Type)
	assert.NotEmpty(t, content.Body)
	assert.False(t, content.CreatedAt.IsZero())
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

	// Create request and response
	req := &colly.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/test",
		},
		Ctx: &colly.Context{},
	}

	resp := &colly.Response{
		Request: req,
		Ctx:     req.Ctx,
	}

	e := &colly.HTMLElement{
		Request:  req,
		Response: resp,
		DOM:      doc.Selection,
	}

	// Test metadata extraction
	metadata := service.ExtractMetadata(e)
	assert.NotNil(t, metadata)
	assert.Equal(t, "article", metadata["type"])
	assert.Equal(t, "Test Title", metadata["title"])
	assert.Equal(t, "Test Description", metadata["description"])
}

func TestDetermineContentType(t *testing.T) {
	testCases := []struct {
		name       string
		url        string
		metadata   map[string]any
		jsonLDType string
		expected   string
	}{
		{
			name:       "uses JSON-LD type",
			url:        "https://example.com/post",
			metadata:   map[string]any{},
			jsonLDType: "Article",
			expected:   "article",
		},
		{
			name:       "uses metadata type",
			url:        "https://example.com/post",
			metadata:   map[string]any{"type": "BlogPost"},
			jsonLDType: "",
			expected:   "blogpost",
		},
		{
			name:       "detects category from URL",
			url:        "https://example.com/category/tech",
			metadata:   map[string]any{},
			jsonLDType: "",
			expected:   "category",
		},
		{
			name:       "defaults to webpage",
			url:        "https://example.com/post",
			metadata:   map[string]any{},
			jsonLDType: "",
			expected:   "webpage",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := content.NewService(nil)
			result := svc.DetermineContentType(tc.url, tc.metadata, tc.jsonLDType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestService_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	svc := content.NewService(mockLogger)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes HTML tags",
			input:    "<p>Hello <b>world</b></p>",
			expected: "Hello world",
		},
		{
			name:     "normalizes whitespace",
			input:    "  Hello   world  \n\t ",
			expected: "Hello world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger.EXPECT().Debug("Processing content", "input", tc.input)
			mockLogger.EXPECT().Debug("Processed content", "result", tc.expected)
			result := svc.Process(t.Context(), tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestService_ProcessBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	svc := content.NewService(mockLogger)

	input := []string{"<p>Hello</p>", "<div>World</div>"}
	expected := []string{"Hello", "World"}

	for i := range input {
		mockLogger.EXPECT().Debug("Processing content", "input", input[i])
		mockLogger.EXPECT().Debug("Processed content", "result", expected[i])
	}

	result := svc.ProcessBatch(t.Context(), input)
	assert.Equal(t, expected, result)
}

func TestService_ProcessWithMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	svc := content.NewService(mockLogger)

	input := "<p>Hello world</p>"
	metadata := map[string]string{
		"source": "test",
		"type":   "article",
	}

	mockLogger.EXPECT().Debug("Processing content with metadata",
		"source", metadata["source"],
		"type", metadata["type"],
	)
	mockLogger.EXPECT().Debug("Processing content", "input", input)
	mockLogger.EXPECT().Debug("Processed content", "result", "Hello world")

	result := svc.ProcessWithMetadata(t.Context(), input, metadata)
	assert.Equal(t, "Hello world", result)
}
