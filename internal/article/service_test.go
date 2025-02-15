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
	svc := article.NewService(mockLogger).(*article.Service)

	tests := []struct {
		name     string
		html     string
		expected time.Time
	}{
		{
			name:     "JSON-LD date",
			html:     `<script type="application/ld+json">{"datePublished": "2025-01-30T23:00:00Z"}</script>`,
			expected: time.Date(2025, 1, 30, 23, 0, 0, 0, time.UTC),
		},
		{
			name:     "timeago date",
			html:     `<time class="timeago" datetime="2025-02-09T16:01:52.203Z">about 23 hours ago</time>`,
			expected: time.Date(2025, 2, 9, 16, 1, 52, 203000000, time.UTC),
		},
		{
			name:     "meta tag date",
			html:     `<meta property="article:published_time" content="2025-02-10T12:00:00Z"/>`,
			expected: time.Date(2025, 2, 10, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			e := &colly.HTMLElement{
				DOM: doc.Selection,
			}

			result := svc.ParsePublishedDate(e, article.JSONLDArticle{})
			assert.Equal(t, tt.expected, result)
		})
	}
}
