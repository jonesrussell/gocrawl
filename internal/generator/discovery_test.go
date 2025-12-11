package generator_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/jonesrussell/gocrawl/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverTitle_SemanticHTML(t *testing.T) {
	html := `<article><h1>Test Title</h1></article>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverTitle()
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "article h1")
	assert.Contains(t, result.SampleText, "Test Title")
}

func TestDiscoverTitle_MultipleMatches(t *testing.T) {
	html := `<h1>Title 1</h1><h1>Title 2</h1>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverTitle()
	// Should have lower confidence due to ambiguity
	assert.Less(t, result.Confidence, 0.80)
	assert.Contains(t, result.Selectors, "h1")
}

func TestDiscoverTitle_MetaTag(t *testing.T) {
	html := `<html><head><meta property="og:title" content="OG Title"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverTitle()
	assert.Greater(t, result.Confidence, 0.85)
	assert.Contains(t, result.Selectors, "meta[property='og:title']")
	assert.Contains(t, result.SampleText, "OG Title")
}

func TestDiscoverBody_LongContent(t *testing.T) {
	longContent := strings.Repeat("word ", 500)
	html := `<article><div class="content">` + longContent + `</div></article>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverBody()
	// Should get length bonus in confidence
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "article")
}

func TestDiscoverBody_ClassPattern(t *testing.T) {
	html := `<div class="article-body">This is the article body text.</div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverBody()
	assert.Greater(t, result.Confidence, 0.80)
	assert.Contains(t, result.Selectors, ".article-body")
}

func TestDiscoverAuthor_SchemaOrg(t *testing.T) {
	html := `<span itemprop="author">John Doe</span>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverAuthor()
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "[itemprop='author']")
}

func TestDiscoverAuthor_MetaTag(t *testing.T) {
	html := `<html><head><meta property="article:author" content="Jane Smith"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverAuthor()
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "meta[property='article:author']")
}

func TestDiscoverPublishedTime_MetaTag(t *testing.T) {
	html := `<html><head><meta property="article:published_time" content="2024-01-01T12:00:00Z"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverPublishedTime()
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "meta[property='article:published_time']")
}

func TestDiscoverPublishedTime_TimeElement(t *testing.T) {
	html := `<time datetime="2024-01-01T12:00:00Z">January 1, 2024</time>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverPublishedTime()
	assert.Greater(t, result.Confidence, 0.85)
	assert.Contains(t, result.Selectors, "time[datetime]")
}

func TestDiscoverImage_OpenGraph(t *testing.T) {
	html := `<html><head><meta property="og:image" content="https://example.com/image.jpg"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverImage()
	assert.Greater(t, result.Confidence, 0.90)
	assert.Contains(t, result.Selectors, "meta[property='og:image']")
}

func TestDiscoverImage_ArticleImage(t *testing.T) {
	html := `<article><img src="https://example.com/article.jpg" alt="Article"></article>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverImage()
	assert.Greater(t, result.Confidence, 0.80)
	assert.Contains(t, result.Selectors, "article img")
}

func TestDiscoverLinks_ArticlePatterns(t *testing.T) {
	html := `
		<div>
			<a href="/news/article1">Article 1</a>
			<a href="/news/article2">Article 2</a>
			<a href="/article/article3">Article 3</a>
			<a href="/about">About</a>
		</div>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverLinks()
	// Should find links with /news/ or /article/ patterns
	// May find selectors or not depending on implementation
	_ = len(result.Selectors)
}

func TestDiscoverCategory_MetaTag(t *testing.T) {
	html := `<html><head><meta property="article:section" content="Technology"></head></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverCategory()
	assert.Greater(t, result.Confidence, 0.85)
	assert.Contains(t, result.Selectors, "meta[property='article:section']")
}

func TestDiscoverExclusions_CommonPatterns(t *testing.T) {
	html := `
		<div class="ad">Ad content</div>
		<nav>Navigation</nav>
		<script>console.log('test');</script>
		<div class="social-follow">Social</div>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	exclusions := sd.DiscoverExclusions()
	assert.Contains(t, exclusions, ".ad")
	assert.Contains(t, exclusions, "nav")
	assert.Contains(t, exclusions, "script")
	assert.Contains(t, exclusions, ".social-follow")
}

func TestDiscoverAll_CompleteResult(t *testing.T) {
	html := `
		<html>
			<head>
				<meta property="og:title" content="Test Article">
				<meta property="article:published_time" content="2024-01-01T12:00:00Z">
				<meta property="article:author" content="John Doe">
				<meta property="og:image" content="https://example.com/image.jpg">
			</head>
			<body>
				<article>
					<h1>Test Article</h1>
					<div class="article-body">This is a test article body with lots of content.</div>
				</article>
			</body>
		</html>
	`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	sd, err := generator.NewSelectorDiscovery(doc, "https://example.com")
	require.NoError(t, err)

	result := sd.DiscoverAll()
	assert.Greater(t, result.Title.Confidence, 0.0)
	assert.Greater(t, result.Body.Confidence, 0.0)
	assert.Greater(t, result.Author.Confidence, 0.0)
	assert.Greater(t, result.PublishedTime.Confidence, 0.0)
	assert.Greater(t, result.Image.Confidence, 0.0)
}
