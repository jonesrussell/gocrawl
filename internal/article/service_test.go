package article_test

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	mockLogger := logger.NewMockLogger() // Assuming you have a mock logger
	svc := article.NewService(mockLogger)

	assert.NotNil(t, svc)
}

func TestParsePublishedDate(t *testing.T) {
	mockLogger := &logger.MockLogger{}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	svc := article.NewService(mockLogger).(*article.Service) // Type assertion to access internal methods

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
			require.NoError(t, err)

			e := &colly.HTMLElement{
				DOM: doc.Selection,
			}

			expectedTime, _ := time.Parse(time.RFC3339, tt.expected)
			result := svc.ParsePublishedDate(e, article.JSONLDArticle{}) // Update to use JSONLDArticle
			assert.Equal(t, expectedTime, result)
		})
	}
}
