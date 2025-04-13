package content_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/gocrawl/internal/content"
	"github.com/jonesrussell/gocrawl/testutils/mocks/logger"
	"github.com/jonesrussell/gocrawl/testutils/mocks/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest creates test dependencies and returns the content service.
func setupTest(
	t *testing.T,
) content.Interface {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockLogger := logger.NewMockInterface(ctrl)
	mockStorage := storage.NewMockInterface(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	return content.NewService(mockLogger, mockStorage)
}

func TestExtractContent(t *testing.T) {
	svc := setupTest(t)

	// Create a test HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body><p>Test content</p></body></html>"))
	require.NoError(t, err)

	// Create a test HTML element
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
		DOM:      doc.Find("p").First(),
	}

	contentData := svc.ExtractContent(e)
	assert.Equal(t, "Test content", contentData)
}

func TestExtractMetadata(t *testing.T) {
	service := setupTest(t)

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
		DOM:      doc.Find("meta[property='article:published_time']").First(),
	}

	metadata := service.ExtractMetadata(e)
	assert.NotNil(t, metadata)
	assert.Equal(t, "2023-04-13T12:00:00Z", metadata["published_time"])
}

func TestService_Process(t *testing.T) {
	svc := setupTest(t)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "<p>Hello world</p>",
			expected: "Hello world",
		},
		{
			name:     "nested elements",
			input:    "<div><p>Hello</p><p>world</p></div>",
			expected: "Hello world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := svc.Process(t.Context(), tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestService_ProcessBatch(t *testing.T) {
	svc := setupTest(t)

	input := []string{"<p>Hello</p>", "<div>World</div>"}
	expected := []string{"Hello", "World"}

	result := svc.ProcessBatch(t.Context(), input)
	assert.Equal(t, expected, result)
}

func TestService_ProcessWithMetadata(t *testing.T) {
	svc := setupTest(t)

	input := "<p>Hello world</p>"
	metadata := map[string]string{
		"title": "Test Title",
		"date":  "2023-04-13",
	}

	result := svc.ProcessWithMetadata(t.Context(), input, metadata)
	assert.Equal(t, "Hello world", result)
}

func TestNewService(t *testing.T) {
	service := setupTest(t)
	assert.NotNil(t, service)
	assert.Implements(t, (*content.Interface)(nil), service)
}

func TestService_ProcessContent(t *testing.T) {
	svc := setupTest(t)

	input := "<p>Test content</p>"
	result := svc.Process(t.Context(), input)
	assert.Equal(t, "Test content", result)
}
