// Package content_test provides tests for the content package.
package content_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	content "github.com/jonesrussell/gocrawl/internal/page"
	"github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"github.com/jonesrussell/gocrawl/testutils/mocks/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCases defines test cases for content processing
type testCase struct {
	name     string
	input    string
	expected string
}

// htmlTestCases defines test cases for HTML content processing
var htmlTestCases = []testCase{
	{
		name:     "simple text",
		input:    `<p>Hello world</p>`,
		expected: "Hello world",
	},
	{
		name:     "nested elements",
		input:    `<div><p>Hello</p><p>world</p></div>`,
		expected: "Hello world",
	},
	{
		name:     "multiple spaces",
		input:    `<p>Hello    world</p>`,
		expected: "Hello world",
	},
	{
		name:     "empty elements",
		input:    `<p></p><p>Hello world</p><p></p>`,
		expected: "Hello world",
	},
	{
		name:     "mixed content",
		input:    `<div><p>Hello</p><span>world</span></div>`,
		expected: "Hello world",
	},
	{
		name:     "whitespace handling",
		input:    `<p>  Hello  world  </p>`,
		expected: "Hello world",
	},
}

// TestService_Process tests the Process method of the Service
func TestService_Process(t *testing.T) {
	// Setup
	service := createTestService(t)

	// Run test cases
	for _, tc := range htmlTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute
			result := service.Process(t.Context(), tc.input)

			// Verify
			assert.Equal(t, tc.expected, result, "Processed content should match expected output")
		})
	}
}

// TestService_ProcessBatch tests the ProcessBatch method of the Service
func TestService_ProcessBatch(t *testing.T) {
	// Setup
	service := createTestService(t)

	// Test cases
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty batch",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single item",
			input:    []string{`<p>Hello world</p>`},
			expected: []string{"Hello world"},
		},
		{
			name: "multiple items",
			input: []string{
				`<p>Hello</p>`,
				`<p>world</p>`,
			},
			expected: []string{"Hello", "world"},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result := service.ProcessBatch(t.Context(), tt.input)

			// Verify
			assert.Equal(t, tt.expected, result, "Processed batch should match expected output")
		})
	}
}

// TestService_ProcessWithMetadata tests the ProcessWithMetadata method of the Service
func TestService_ProcessWithMetadata(t *testing.T) {
	// Setup
	service := createTestService(t)

	// Test cases
	tests := []struct {
		name     string
		input    string
		metadata map[string]string
		expected string
	}{
		{
			name:     "no metadata",
			input:    `<p>Hello world</p>`,
			metadata: nil,
			expected: "Hello world",
		},
		{
			name:     "with metadata",
			input:    `<p>Hello world</p>`,
			metadata: map[string]string{"source": "test", "type": "article"},
			expected: "Hello world",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result := service.ProcessWithMetadata(t.Context(), tt.input, tt.metadata)

			// Verify
			assert.Equal(t, tt.expected, result, "Processed content with metadata should match expected output")
		})
	}
}

// createTestService creates a test service instance
func createTestService(t *testing.T) content.Interface {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockInterface(ctrl)
	mockStorage := storage.NewMockInterface(ctrl)

	// Set up default expectations
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	return content.NewService(mockLogger, mockStorage)
}

// createTestDocument creates a test goquery Document
func createTestDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err, "Failed to create test document")
	return doc
}

// createTestHTMLElement creates a test colly HTMLElement
func createTestHTMLElement(t *testing.T, html string) *colly.HTMLElement {
	t.Helper()
	doc := createTestDocument(t, html)
	return &colly.HTMLElement{
		DOM: doc.Find("body"),
		Request: &colly.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		},
	}
}

func TestExtractContent(t *testing.T) {
	svc := createTestService(t)

	// Create a test HTML element with a body element
	html := `<body><p>Test content</p></body>`
	e := createTestHTMLElement(t, html)

	// Extract content
	content := svc.ExtractContent(e)

	// Verify content
	assert.NotNil(t, content)
	assert.Equal(t, "Test content", content.Body)
	assert.Equal(t, "http://example.com/test", content.URL)
}

func TestExtractMetadata(t *testing.T) {
	service := createTestService(t)

	// Create a test HTML element
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html>
			<head>
				<meta property="article:published_time" content="2023-04-13T12:00:00Z">
			</head>
			<body>
				<p>Test content</p>
			</body>
		</html>
	`))
	require.NoError(t, err)

	req := &colly.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/test",
		},
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

	metadata := service.ExtractMetadata(e)
	assert.NotNil(t, metadata)
	assert.Equal(t, "2023-04-13T12:00:00Z", metadata["article:published_time"])
}

func TestNewService(t *testing.T) {
	service := createTestService(t)
	assert.NotNil(t, service)
	assert.Implements(t, (*content.Interface)(nil), service)
}

func TestService_ProcessContent(t *testing.T) {
	svc := createTestService(t)

	input := "<p>Test content</p>"
	result := svc.Process(t.Context(), input)
	assert.Equal(t, "Test content", result)
}
