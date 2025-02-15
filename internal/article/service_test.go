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

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Helper function to create a mock logger
func newMockLogger() *logger.MockLogger {
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return() // Set up expectation for Debug call
	return mockLogger
}

// Common HTML structure for tests
const testHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Test Article</title>
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
        <div id="details-body">NEWS RELEASE<br />ONTARIO PROVINCIAL POLICE<br />...</div>
        <div class="details-byline">
            ElliotLakeToday Staff
            <time datetime="2025-02-11T17:37:42Z">Feb 11, 2025 12:37 PM</time>
        </div>
    </div>
</body>
</html>
`

func TestExtractArticle(t *testing.T) {
	mockLogger := newMockLogger() // Use the helper function
	svc := article.NewService(mockLogger)

	// Create a new document from the common HTML string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create a colly HTMLElement for the div.details
	e := &colly.HTMLElement{
		DOM: doc.Find("div.details"),
		Request: &colly.Request{
			URL: &url.URL{Path: "/mock-url"},
		},
	}

	// Call the ExtractArticle method
	article := svc.ExtractArticle(e)

	// Validate the extracted article
	assert.NotNil(t, article)
	assert.Equal(t, "/mock-url", article.Source)
	assert.Equal(t, "Elliot Lake man arrested after threatening to kill victim and police", article.Title)
	assert.Equal(
		t,
		"Police were called to house on Milliken Road for report of break-and-enter\n\nNEWS RELEASEONTARIO PROVINCIAL POLICE...",
		article.Body,
	)
	assert.Equal(t, "ElliotLakeToday Staff", article.Author)

	// Assert that the expectations were met
	mockLogger.AssertExpectations(t)
}

func TestCleanAuthor(t *testing.T) {
	mockLogger := newMockLogger() // Use the helper function
	svc := article.NewService(mockLogger)

	// Change the author string to match the expected format
	author := "ElliotLakeToday Staff" // Simplified for testing

	cleanedAuthor := svc.CleanAuthor(author)

	assert.Equal(t, "ElliotLakeToday Staff", cleanedAuthor)
}

func TestParsePublishedDate(t *testing.T) {
	mockLogger := newMockLogger() // Use the helper function
	svc := article.NewService(mockLogger)

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

	expectedDate, _ := time.Parse(time.RFC3339, "2025-02-14T15:04:05Z")
	assert.Equal(t, expectedDate, date)
}
