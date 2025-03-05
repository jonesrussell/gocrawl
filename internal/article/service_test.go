package article_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Helper function to create a mock logger
func newMockLogger() *logger.MockLogger {
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return() // Set up expectation for Debug call
	return mockLogger
}

// Helper function to create default article selectors
func newDefaultSelectors() config.ArticleSelectors {
	return config.ArticleSelectors{
		Container:     "div.details",
		Title:         "h1.details-title",
		Body:          "div.details-body",
		Intro:         "div.details-intro",
		Byline:        "div.details-byline",
		Author:        "span.author",
		PublishedTime: "time.timeago",
		TimeAgo:       "time.timeago",
		JSONLD:        "script[type='application/ld+json']",
	}
}

// Common HTML structure for tests
const testHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Test Article</title>
    <meta property="og:type" content="article">
    <script type="application/ld+json">
    {
        "datePublished": "2025-02-11T17:37:42Z",
        "author": "ElliotLakeToday Staff",
        "keywords": ["OPP", "arrest", "assault", "Elliot Lake"],
        "articleSection": "Police"
    }
    </script>
</head>
<body>
    <div class="details">
        <h1 class="details-title">Elliot Lake man arrested after threatening to kill victim and police</h1>
        <div class="details-intro">Police were called to house on Milliken Road for report of break-and-enter</div>
        <div class="details-byline">
            <span class="author">ElliotLakeToday Staff</span>
            <time class="timeago" datetime="2025-02-11T17:37:42Z">Feb 11, 2025 12:37 PM</time>
        </div>
        <div class="details-body">
            Police were called to house on Milliken Road for report of break-and-enter
        </div>
    </div>
</body>
</html>
`

func TestExtractArticle(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	selectors := newDefaultSelectors()
	service := article.NewService(mockLogger, selectors)

	// Set up mock expectations for all debug calls with variable arguments
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return()
	mockLogger.On("Debug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return()
	mockLogger.On("Debug", "No valid published date found", "dates", mock.AnythingOfType("[]string")).Return()
	mockLogger.On("Debug", "Successfully parsed date",
		"source", mock.AnythingOfType("string"),
		"format", mock.AnythingOfType("string"),
		"result", mock.AnythingOfType("time.Time")).Return()
	mockLogger.On("Debug", "Extracted article",
		"component", "article/service",
		"id", mock.AnythingOfType("string"),
		"title", mock.AnythingOfType("string"),
		"url", mock.AnythingOfType("string"),
		"date", mock.AnythingOfType("time.Time"),
		"author", mock.AnythingOfType("string"),
		"tags", mock.AnythingOfType("[]string"),
		"wordCount", mock.AnythingOfType("int"),
		"category", mock.AnythingOfType("string"),
		"section", mock.AnythingOfType("string")).Return()

	// Create a new document from the common HTML string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create a colly HTMLElement for the div.details
	e := &colly.HTMLElement{
		DOM: doc.Find(selectors.Container),
		Request: &colly.Request{
			URL: &url.URL{Path: "/mock-url"},
		},
	}

	// Call the ExtractArticle method
	article := service.ExtractArticle(e)

	// Validate the extracted article
	require.NotNil(t, article)
	require.Equal(t, "/mock-url", article.Source)
	require.Equal(t, "Elliot Lake man arrested after threatening to kill victim and police", article.Title)
	require.Equal(t, "Police were called to house on Milliken Road for report of break-and-enter", strings.TrimSpace(article.Body))
	require.Equal(t, "ElliotLakeToday Staff", article.Author)
}

func TestCleanAuthor(t *testing.T) {
	mockLogger := newMockLogger()
	svc := article.NewService(mockLogger, newDefaultSelectors())

	// Change the author string to match the expected format
	author := "ElliotLakeToday Staff" // Simplified for testing

	cleanedAuthor := svc.CleanAuthor(author)

	assert.Equal(t, "ElliotLakeToday Staff", cleanedAuthor)
}

func TestParsePublishedDate(t *testing.T) {
	mockLogger := logger.NewMockLogger()
	svc := article.NewService(mockLogger, newDefaultSelectors())

	// Set up mock expectations for debug calls
	mockLogger.On("Debug",
		"Trying to parse date",
		"value", "2025-02-14T15:04:05Z",
	).Return()

	expectedDate, _ := time.Parse(time.RFC3339, "2025-02-14T15:04:05Z")
	mockLogger.On("Debug",
		"Successfully parsed date",
		"source", "2025-02-14T15:04:05Z",
		"format", "2006-01-02T15:04:05Z07:00",
		"result", expectedDate,
	).Return()

	// Create a mock HTML document
	html := `
        <html>
            <head>
                <script type="application/ld+json">{"datePublished": "2025-02-14T15:04:05Z"}</script>
            </head>
            <body>
                <time datetime="2025-02-14T15:04:05Z">about 23 hours ago</time>
            </body>
        </html>
    `

	// Create a new document from the HTML string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create a colly HTMLElement with the mock document
	e := &colly.HTMLElement{
		DOM: doc.Selection,
		Request: &colly.Request{
			URL: &url.URL{Path: "/mock-url"},
		},
	}

	jsonLD := article.JSONLDArticle{
		DatePublished: "2025-02-14T15:04:05Z",
	}

	date := svc.ParsePublishedDate(e, jsonLD)

	assert.Equal(t, expectedDate, date)
}
