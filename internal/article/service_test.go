package article

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestArticleService(t *testing.T) {
	mockLogger := logger.NewMockCustomLogger()
	svc := &ArticleService{logger: mockLogger}

	t.Run("CleanAuthor", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "with date",
				input:    "John Doe    Feb 10, 2024",
				expected: "John Doe",
			},
			{
				name:     "without date",
				input:    "Jane Smith",
				expected: "Jane Smith",
			},
			{
				name:     "empty string",
				input:    "",
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := svc.CleanAuthor(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("ParsePublishedDate", func(t *testing.T) {
		e := &colly.HTMLElement{} // Mock HTMLElement
		jsonLD := jsonLDArticle{
			DatePublished: "2024-02-10T16:01:52Z",
		}

		result := svc.ParsePublishedDate(e, jsonLD)
		expected := time.Date(2024, 2, 10, 16, 1, 52, 0, time.UTC)
		assert.Equal(t, expected, result)
	})
}

func TestExtractArticle(t *testing.T) {
	// Setup
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	svc := NewService(mockLogger)

	t.Run("empty article", func(t *testing.T) {
		e := &colly.HTMLElement{} // Mock HTMLElement
		article := svc.ExtractArticle(e)
		assert.Nil(t, article)
	})

	t.Run("full article", func(t *testing.T) {
		// Create test HTML
		html := `
			<html>
				<head>
					<script type="application/ld+json">
						{
							"dateCreated": "2025-01-30T20:17:59Z",
							"dateModified": "2025-01-30T23:00:00Z",
							"datePublished": "2025-01-30T23:00:00Z",
							"author": "ElliotLakeToday Staff",
							"keywords": ["OPP", "fraud", "RCMP"],
							"articleSection": "Police"
						}
					</script>
				</head>
				<body>
					<h1 class="details-title">Test Article Title</h1>
					<div class="details-intro">Test intro text</div>
					<div id="details-body">Test article body</div>
					<div class="details-byline">John Doe     Feb 10, 2025</div>
					<time class="timeago" datetime="2025-02-09T16:01:52.203Z">about 23 hours ago</time>
					<meta property="article:section" content="News"/>
					<meta name="keywords" content="test|article|keywords"/>
				</body>
			</html>
		`

		// Create HTMLElement using goquery
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		assert.NoError(t, err)

		testURL, _ := url.Parse("https://www.elliotlaketoday.com/opp-beat/test-article")
		e := &colly.HTMLElement{
			DOM: doc.Selection,
			Request: &colly.Request{
				URL: testURL,
			},
			Response: &colly.Response{
				Request: &colly.Request{URL: testURL},
			},
		}

		// Act
		article := svc.ExtractArticle(e)

		// Assert
		assert.NotNil(t, article)
		assert.Equal(t, "Test Article Title", article.Title)
		assert.Equal(t, "Test intro text\n\nTest article body", article.Body)
		assert.Equal(t, "John Doe", article.Author)
		assert.Equal(t, "https://www.elliotlaketoday.com/opp-beat/test-article", article.Source)

		expectedTime, _ := time.Parse(time.RFC3339, "2025-01-30T23:00:00Z")
		assert.Equal(t, expectedTime, article.PublishedDate)

		// Check tags
		expectedTags := []string{"OPP", "fraud", "RCMP", "Police", "News", "test", "article", "keywords", "OPP Beat"}
		assert.ElementsMatch(t, expectedTags, article.Tags)
	})

	mockLogger.AssertExpectations(t)
}

func TestCleanAuthor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "author with date",
			input:    "John Doe     Feb 10, 2025",
			expected: "John Doe",
		},
		{
			name:     "author without date",
			input:    "Jane Smith",
			expected: "Jane Smith",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	svc := NewService(mockLogger).(*ArticleService) // Type assertion to access internal methods

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CleanAuthor(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePublishedDate(t *testing.T) {
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	svc := NewService(mockLogger).(*ArticleService) // Type assertion to access internal methods

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "JSON-LD date",
			html:     `<script type="application/ld+json">{"datePublished": "2025-01-30T23:00:00Z"}</script>`,
			expected: "2025-01-30T23:00:00Z",
		},
		{
			name:     "timeago date",
			html:     `<time class="timeago" datetime="2025-02-09T16:01:52.203Z">about 23 hours ago</time>`,
			expected: "2025-02-09T16:01:52.203Z",
		},
		{
			name:     "meta tag date",
			html:     `<meta property="article:published_time" content="2025-02-10T12:00:00Z"/>`,
			expected: "2025-02-10T12:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err)

			e := &colly.HTMLElement{
				DOM: doc.Selection,
			}

			expectedTime, _ := time.Parse(time.RFC3339, tt.expected)
			result := svc.ParsePublishedDate(e, jsonLDArticle{})
			assert.Equal(t, expectedTime, result)
		})
	}
}
