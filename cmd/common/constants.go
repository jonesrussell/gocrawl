package common

import "time"

const (
	// DefaultCrawlerTimeout is the maximum time to wait for the crawler to complete.
	DefaultCrawlerTimeout = 30 * time.Minute

	// DefaultShutdownTimeout is the maximum time to wait for graceful shutdown.
	DefaultShutdownTimeout = 30 * time.Second
)

// Index name constants
const (
	// DefaultArticleIndex is the default index name for articles
	DefaultArticleIndex = "articles"

	// DefaultPageIndex is the default index name for pages
	DefaultPageIndex = "pages"

	// DefaultContentIndex is the default index name for general content
	DefaultContentIndex = "content"
)

// Capacity constants
const (
	// DefaultIndicesCapacity is the initial capacity for index slices.
	// Set to 2 to accommodate both content and article indices for a source.
	DefaultIndicesCapacity = 2
)
