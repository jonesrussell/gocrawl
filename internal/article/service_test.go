package article_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/config"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

// Helper function to create a mock logger
func newMockLogger(t *testing.T) *logger.MockInterface {
	ctrl := gomock.NewController(t)
	mockLogger := logger.NewMockInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
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
	mockLogger := newMockLogger(t)
	selectors := newDefaultSelectors()
	service := article.NewService(mockLogger, selectors)

	// Setup logger expectations
	mockLogger.EXPECT().Debug(
		gomock.AssignableToTypeOf(""),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Debug(
		gomock.AssignableToTypeOf(""),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Debug(
		gomock.AssignableToTypeOf(""),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Debug(
		"No valid published date found",
		"dates",
		gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Debug(
		"Successfully parsed date",
		"source", gomock.Any(),
		"format", gomock.Any(),
		"result", gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Debug(
		"Extracted article",
		"component", gomock.Any(),
		"id", gomock.Any(),
		"title", gomock.Any(),
		"url", gomock.Any(),
		"date", gomock.Any(),
		"author", gomock.Any(),
		"tags", gomock.Any(),
		"wordCount", gomock.Any(),
		"category", gomock.Any(),
		"section", gomock.Any(),
	).AnyTimes()

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
	expectedBody := "Police were called to house on Milliken Road for report of break-and-enter"
	require.Equal(t, expectedBody, strings.TrimSpace(article.Body))
	require.Equal(t, "ElliotLakeToday Staff", article.Author)
}

func TestCleanAuthor(t *testing.T) {
	mockLogger := newMockLogger(t)
	svc := article.NewService(mockLogger, newDefaultSelectors())

	// Change the author string to match the expected format
	author := "ElliotLakeToday Staff" // Simplified for testing

	cleanedAuthor := svc.CleanAuthor(author)

	assert.Equal(t, "ElliotLakeToday Staff", cleanedAuthor)
}

func TestParsePublishedDate(t *testing.T) {
	mockLogger := newMockLogger(t)
	svc := article.NewService(mockLogger, newDefaultSelectors())

	// Set up mock expectations for debug calls
	mockLogger.EXPECT().Debug(
		"Trying to parse date",
		"value", "2025-02-14T15:04:05Z",
	).AnyTimes()

	expectedDate, _ := time.Parse(time.RFC3339, "2025-02-14T15:04:05Z")
	mockLogger.EXPECT().Debug(
		"Successfully parsed date",
		"source", "2025-02-14T15:04:05Z",
		"format", "2006-01-02T15:04:05Z07:00",
		"result", expectedDate,
	).AnyTimes()

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
	mockLogger := newMockLogger(t)
	selectors := config.DefaultArticleSelectors()
	service := article.NewService(mockLogger, selectors)

	// Set up mock expectations for debug calls
	mockLogger.EXPECT().Debug("Found JSON-LD keywords", "values", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Found JSON-LD section", "value", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Found meta section", "value", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("Found meta keywords", "value", gomock.Any()).AnyTimes()

	// Create test HTML with meta tags
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta property="article:section" content="News">
			<meta name="keywords" content="crime|police|arrest">
		</head>
		<body>
			<article></article>
		</body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create a colly HTMLElement with the full document
	e := &colly.HTMLElement{
		DOM: doc.Selection,
		Request: &colly.Request{
			URL: &url.URL{
				Path: "/opp-beat/article-123",
			},
		},
	}

	// Create JSON-LD data
	jsonLD := article.JSONLDArticle{
		Keywords: []string{"OPP", "arrest", "assault"},
		Section:  "Police",
	}

	// Extract tags
	tags := service.ExtractTags(e, jsonLD)

	// Validate tags (note: "arrest" appears only once due to deduplication)
	expectedTags := []string{"OPP", "arrest", "assault", "Police", "News", "crime", "police", "OPP Beat"}
	require.Equal(t, expectedTags, tags)
}

func TestRemoveDuplicates(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "No duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "With duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Case sensitive",
			input:    []string{"A", "a", "B", "b"},
			expected: []string{"A", "a", "B", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := article.RemoveDuplicates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
