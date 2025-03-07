// Package collector provides the web page collection functionality for GoCrawl.
// It manages the actual crawling process using the Colly web scraping framework,
// handling URL processing, rate limiting, and content extraction.
package collector

// Constants for default selectors used in content extraction.
// These selectors are used when no specific selectors are provided in the source configuration.
const (
	// DefaultArticleSelector is the default CSS selector for finding article containers.
	// It looks for elements with the 'article' tag or class.
	DefaultArticleSelector = "article, .article"

	// DefaultTitleSelector is the default CSS selector for finding article titles.
	// It looks for h1 and h2 heading elements.
	DefaultTitleSelector = "h1, h2"

	// DefaultDateSelector is the default CSS selector for finding publication dates.
	// It looks for elements with the 'time' tag.
	DefaultDateSelector = "time"

	// DefaultAuthorSelector is the default CSS selector for finding author information.
	// It looks for elements with the 'author' class.
	DefaultAuthorSelector = ".author"

	// DefaultCategoriesSelector is the default CSS selector for finding category information.
	// It looks for elements with the 'categories' class within a div.
	DefaultCategoriesSelector = "div.categories"
)
