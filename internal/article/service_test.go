package article_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/gocrawl/internal/article"
	"github.com/jonesrussell/gocrawl/internal/logger"
)

func TestExtractArticle(t *testing.T) {
	svc := article.NewService(&logger.MockCustomLogger{})
	e := &colly.HTMLElement{
		Request: &colly.Request{URL: &url.URL{Path: "/mock-url"}},
	}

	article := svc.ExtractArticle(e)

	assert.NotNil(t, article)
	assert.NotEmpty(t, article.ID)
	assert.Equal(t, "/mock-url", article.Source)
}

func TestCleanAuthor(t *testing.T) {
	svc := article.NewService(&logger.MockCustomLogger{})
	author := "John Doe    Feb 14, 2025"

	cleanedAuthor := svc.CleanAuthor(author)

	assert.Equal(t, "John Doe", cleanedAuthor)
}

func TestExtractTags(t *testing.T) {
	svc := article.NewService(&logger.MockCustomLogger{})
	e := &colly.HTMLElement{
		Request: &colly.Request{URL: &url.URL{Path: "/opp-beat/"}},
	}
	jsonLD := article.JSONLDArticle{
		Keywords: []string{"tag1", "tag2"},
		Section:  "news",
	}

	tags := svc.ExtractTags(e, jsonLD)

	expectedTags := []string{"tag1", "tag2", "news", "OPP Beat"}
	assert.ElementsMatch(t, expectedTags, tags)
}

func TestParsePublishedDate(t *testing.T) {
	svc := article.NewService(&logger.MockCustomLogger{})
	e := &colly.HTMLElement{}
	jsonLD := article.JSONLDArticle{
		DatePublished: "2025-02-14T15:04:05Z",
	}

	date := svc.ParsePublishedDate(e, jsonLD)

	expectedDate, _ := time.Parse(time.RFC3339, "2025-02-14T15:04:05Z")
	assert.Equal(t, expectedDate, date)
}

func TestRemoveDuplicates(t *testing.T) {
	tags := []string{"tag1", "tag2", "tag1", "tag3", "tag2"}
	expectedTags := []string{"tag1", "tag2", "tag3"}

	uniqueTags := article.RemoveDuplicates(tags)

	assert.ElementsMatch(t, expectedTags, uniqueTags)
}
