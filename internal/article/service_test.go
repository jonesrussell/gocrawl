package article_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a mock logger
func newMockLogger() *testutils.MockLogger {
	mockLogger := &testutils.MockLogger{}
	mockLogger.On(
		"Debug",
		"Successfully parsed date",
		"source", "2025-02-14T15:04:05Z",
		"format", "2006-01-02T15:04:05Z07:00",
		"result", time.Date(2025, 2, 14, 15, 4, 5, 0, time.UTC),
	).Return()
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
	mockLogger := newMockLogger()
	selectors := newDefaultSelectors()
	service := article.NewService(mockLogger, selectors)

	// Setup logger expectations
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything,
	).Return()
	mockLogger.On("Debug",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything,
	).Return()
	mockLogger.On("Debug", "No valid published date found", "dates", mock.Anything).Return()
	mockLogger.On("Debug",
		"Successfully parsed date",
		"source", mock.Anything,
		"format", mock.Anything,
		"result", mock.Anything,
	).Return()
	mockLogger.On("Debug",
		"Extracted articleData",
		"component", mock.Anything,
		"id", mock.Anything,
		"title", mock.Anything,
		"url", mock.Anything,
		"date", mock.Anything,
		"author", mock.Anything,
		"tags", mock.Anything,
		"wordCount", mock.Anything,
		"category", mock.Anything,
		"section", mock.Anything,
	).Return()

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
	articleData := service.ExtractArticle(e)

	// Validate the extracted articleData
	require.NotNil(t, articleData)
	require.Equal(t, "/mock-url", articleData.Source)
	require.Equal(t, "Elliot Lake man arrested after threatening to kill victim and police", articleData.Title)
	expectedBody := "Police were called to house on Milliken Road for report of break-and-enter"
	require.Equal(t, expectedBody, strings.TrimSpace(articleData.Body))
	require.Equal(t, "ElliotLakeToday Staff", articleData.Author)
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
	mockLogger := newMockLogger()
	svc := article.NewService(mockLogger, newDefaultSelectors())

	// Set up mock expectations for debug calls
	mockLogger.On("Debug", "Trying to parse date", "value", "2025-02-14T15:04:05Z").Return()

	expectedDate, _ := time.Parse(time.RFC3339, "2025-02-14T15:04:05Z")
	mockLogger.On(
		"Debug",
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

func TestExtractTags(t *testing.T) {
	mockLogger := newMockLogger()
	selectors := config.DefaultArticleSelectors()
	service := article.NewService(mockLogger, selectors)

	// Set up mock expectations for debug calls
	mockLogger.On("Debug", "Found JSON-LD keywords", "values", mock.Anything).Return()
	mockLogger.On("Debug", "Found JSON-LD section", "value", mock.Anything).Return()
	mockLogger.On("Debug", "Found meta section", "value", mock.Anything).Return()
	mockLogger.On("Debug", "Found meta keywords", "value", mock.Anything).Return()

	// Create test HTML with meta tags
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta property="article:section" content="News">
			<meta name="keywords" content="crime|police|arrest">
		</head>
		<body>
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

	// Create JSON-LD data
	jsonLD := article.JSONLDArticle{
		Keywords: []string{"OPP", "arrest", "assault"},
		Section:  "Police",
	}

	// Call the ExtractTags method
	tags := service.ExtractTags(e, jsonLD)

	// Validate the extracted tags
	require.NotNil(t, tags)
	require.Contains(t, tags, "crime")
	require.Contains(t, tags, "police")
	require.Contains(t, tags, "arrest")
	require.Contains(t, tags, "News")
	require.Contains(t, tags, "OPP")
	require.Contains(t, tags, "assault")
	require.Contains(t, tags, "Police")
}

func TestRemoveDuplicates(t *testing.T) {
	// Test case with duplicates
	input := []string{"a", "b", "a", "c", "b", "d"}
	expected := []string{"a", "b", "c", "d"}

	result := article.RemoveDuplicates(input)

	assert.Equal(t, expected, result)
}
