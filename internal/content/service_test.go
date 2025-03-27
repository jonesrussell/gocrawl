package content_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExtractContent(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
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
	<meta property="article:published_time" contentData="2024-03-15T10:00:00Z" />
</head>
<body>
	<h1>Test Article</h1>
	<p>Test contentData</p>
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

	// Set up logger expectations
	mockLogger.On("Debug", "Extracting content", "url", testURL).Return()
	mockLogger.On("Debug", "Trying to parse date", "value", "2024-03-15T10:00:00Z").Return()
	mockLogger.On("Debug", "Successfully parsed date",
		"source", "2024-03-15T10:00:00Z",
		"format", time.RFC3339,
		"result", "2024-03-15 10:00:00 +0000 UTC",
	).Return()
	mockLogger.On("Debug", "Extracted content",
		"id", mock.Anything,
		"title", "Test Article",
		"url", testURL,
		"type", "article",
		"created_at", mock.Anything,
	).Return()

	contentData := svc.ExtractContent(e)
	require.NotNil(t, contentData)
	assert.Equal(t, "Test Article", contentData.Title)
	assert.Equal(t, testURL, contentData.URL)
	assert.Equal(t, "article", contentData.Type)
	assert.NotEmpty(t, contentData.Body)
	assert.False(t, contentData.CreatedAt.IsZero())
}

func TestExtractMetadata(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
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
	mockLogger := &testutils.MockLogger{}
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
			mockLogger.On("Debug", "Processing content", "input", tc.input).Return()
			mockLogger.On("Debug", "Processed content", "result", tc.expected).Return()
			result := svc.Process(t.Context(), tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestService_ProcessBatch(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	svc := content.NewService(mockLogger)

	input := []string{"<p>Hello</p>", "<div>World</div>"}
	expected := []string{"Hello", "World"}

	for i := range input {
		mockLogger.On("Debug", "Processing content", "input", input[i]).Return()
		mockLogger.On("Debug", "Processed content", "result", expected[i]).Return()
	}

	result := svc.ProcessBatch(t.Context(), input)
	assert.Equal(t, expected, result)
}

func TestService_ProcessWithMetadata(t *testing.T) {
	mockLogger := &testutils.MockLogger{}
	svc := content.NewService(mockLogger)

	input := "<p>Hello world</p>"
	metadata := map[string]string{
		"source": "test",
		"type":   "article",
	}

	mockLogger.On("Debug", "Processing content with metadata",
		"source", metadata["source"],
		"type", metadata["type"],
	).Return()
	mockLogger.On("Debug", "Processing content", "input", input).Return()
	mockLogger.On("Debug", "Processed content", "result", "Hello world").Return()

	result := svc.ProcessWithMetadata(t.Context(), input, metadata)
	assert.Equal(t, "Hello world", result)
}
