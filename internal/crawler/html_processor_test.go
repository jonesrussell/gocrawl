package crawler_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/config/types"
	"github.com/jonesrussell/gocrawl/internal/constants"
	"github.com/jonesrussell/gocrawl/internal/content/contenttype"
	"github.com/jonesrussell/gocrawl/internal/crawler"
	"github.com/jonesrussell/gocrawl/internal/logger"
	"github.com/jonesrussell/gocrawl/internal/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createHTMLElement creates a colly HTMLElement from HTML string for testing
func createHTMLElement(htmlContent string) (*colly.HTMLElement, error) {
	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Get the html element
	htmlSelection := doc.Find("html")
	if htmlSelection.Length() == 0 {
		// If no html element, create a wrapper
		htmlSelection = doc.Selection
	}

	// Create a mock request
	parsedURL, err := url.Parse("https://example.com/test")
	if err != nil {
		return nil, err
	}

	req := &colly.Request{
		URL:     parsedURL,
		Method:  "GET",
		Headers: &http.Header{},
	}

	resp := &colly.Response{
		Request:    req,
		StatusCode: 200,
		Body:       []byte(htmlContent),
	}

	// Create HTMLElement with the goquery selection
	return &colly.HTMLElement{
		Request:  req,
		Response: resp,
		DOM:      htmlSelection,
	}, nil
}

func TestDetectContentType_Article_ViaMetadata(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html>
<head>
	<meta property="og:type" content="article">
</head>
<body>
	<article>Test content</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Body: "article",
			},
		},
	}

	log := logger.NewNoOp()
	// Create a minimal Sources instance - we can pass nil since we're passing source directly
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Article, contentType)
}

func TestDetectContentType_Article_ViaSelectors(t *testing.T) {
	t.Parallel()

	// Article with substantial content (>200 chars)
	articleBody := strings.Repeat("This is article content. ", 50) // >200 chars

	html := `<!DOCTYPE html>
<html>
<body>
	<article>
		<h1>Article Title</h1>
		<div class="article-body">` + articleBody + `</div>
	</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1",
				Body:  ".article-body",
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Article, contentType)
}

func TestDetectContentType_Page_ShortContent(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html>
<body>
	<article>
		<h1>Page Title</h1>
		<div class="article-body">Short text</div>
	</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1",
				Body:  ".article-body",
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Page, contentType, "Short content should be detected as page")
}

func TestDetectContentType_Page_NoBodySelector(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html>
<body>
	<div class="content">Some content</div>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1",
				Body:  ".article-body", // Won't match
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Page, contentType, "No body match should be detected as page")
}

func TestDetectContentType_Page_NoTitle(t *testing.T) {
	t.Parallel()

	articleBody := strings.Repeat("This is content. ", 50) // >200 chars

	html := `<!DOCTYPE html>
<html>
<body>
	<article>
		<div class="article-body">` + articleBody + `</div>
	</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1", // Won't match - no h1 in HTML
				Body:  ".article-body",
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Page, contentType, "No title should be detected as page")
}

func TestDetectContentType_Page_NoSource(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html>
<body>
	<article>Content</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, nil)
	assert.Equal(t, contenttype.Page, contentType, "No source should default to page")
}

func TestDetectContentType_Page_EmptyBodySelector(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html>
<body>
	<article>Content</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Body: "", // Empty selector
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Page, contentType, "Empty body selector should default to page")
}

func TestDetectContentType_Article_ExactMinLength(t *testing.T) {
	t.Parallel()

	// Create content exactly at the minimum length
	articleBody := strings.Repeat("a", constants.MinArticleBodyLength)

	html := `<!DOCTYPE html>
<html>
<body>
	<article>
		<h1>Article Title</h1>
		<div class="article-body">` + articleBody + `</div>
	</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1",
				Body:  ".article-body",
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Article, contentType, "Content at minimum length should be detected as article")
}

func TestDetectContentType_Page_JustBelowMinLength(t *testing.T) {
	t.Parallel()

	// Create content just below the minimum length
	articleBody := strings.Repeat("a", constants.MinArticleBodyLength-1)

	html := `<!DOCTYPE html>
<html>
<body>
	<article>
		<h1>Article Title</h1>
		<div class="article-body">` + articleBody + `</div>
	</article>
</body>
</html>`

	e, err := createHTMLElement(html)
	require.NoError(t, err)

	source := &types.Source{
		Selectors: types.SourceSelectors{
			Article: types.ArticleSelectors{
				Title: "h1",
				Body:  ".article-body",
			},
		},
	}

	log := logger.NewNoOp()
	sourcesInstance := &sources.Sources{}
	proc := crawler.NewHTMLProcessor(log, sourcesInstance)

	contentType := proc.DetectContentType(e, source)
	assert.Equal(t, contenttype.Page, contentType, "Content just below minimum length should be detected as page")
}
