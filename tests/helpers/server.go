// Package helpers provides testing utilities for integration tests.
package helpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockCrawlTarget creates a mock website for crawling tests.
// It returns a test server that serves the provided content map.
// The map key is the URL path, and the value is the HTML content to serve.
func MockCrawlTarget(content map[string]string) *httptest.Server {
	mux := http.NewServeMux()

	// Default content if none provided
	if len(content) == 0 {
		content = map[string]string{
			"/": TestHTMLPage("Home", "Welcome to the test site"),
		}
	}

	// Serve content from the map
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		html, ok := content[path]
		if ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, html)
		} else {
			// Try to serve default content
			defaultHTML, hasDefault := content["/"]
			if hasDefault {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, defaultHTML)
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprint(w, "<html><body>404 Not Found</body></html>")
			}
		}
	})

	return httptest.NewServer(mux)
}

// StartTestServer starts a test HTTP server serving mock content.
// This is a convenience wrapper around MockCrawlTarget.
func StartTestServer(content map[string]string) *httptest.Server {
	return MockCrawlTarget(content)
}

// MockCrawlTargetWithLinks creates a mock website with linked pages.
// It automatically generates links between pages based on the content map.
func MockCrawlTargetWithLinks(baseURL string, pages map[string]PageContent) *httptest.Server {
	mux := http.NewServeMux()
	mu := &sync.Mutex{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		path := r.URL.Path
		page, ok := pages[path]
		if ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, page.HTML)
		} else if path == "/" {
			// Serve index page
			indexPage, hasIndex := pages["/"]
			if hasIndex {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, indexPage.HTML)
			} else {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprint(w, "<html><body>404 Not Found</body></html>")
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, "<html><body>404 Not Found</body></html>")
		}
	})

	return httptest.NewServer(mux)
}

// PageContent represents content for a single page in a mock website.
type PageContent struct {
	Title string
	Body  string
	HTML  string
	Links []string // Relative URLs to link to
}

// CreateLinkedPages creates a set of linked pages for testing.
func CreateLinkedPages() map[string]PageContent {
	return map[string]PageContent{
		"/": {
			Title: "Home",
			Body:  "Welcome to the test site",
			HTML:  TestHTMLPage("Home", "Welcome to the test site. <a href=\"/page1\">Page 1</a> <a href=\"/page2\">Page 2</a>"),
			Links: []string{"/page1", "/page2"},
		},
		"/page1": {
			Title: "Page 1",
			Body:  "This is page 1",
			HTML:  TestHTMLPage("Page 1", "This is page 1. <a href=\"/\">Home</a>"),
			Links: []string{"/"},
		},
		"/page2": {
			Title: "Page 2",
			Body:  "This is page 2",
			HTML:  TestHTMLPage("Page 2", "This is page 2. <a href=\"/\">Home</a>"),
			Links: []string{"/"},
		},
	}
}

// CreateArticlePages creates a set of article pages for testing.
func CreateArticlePages() map[string]string {
	return map[string]string{
		"/": TestHTMLPage("Home", `
			<h2>Articles</h2>
			<ul>
				<li><a href="/article/1">Article 1</a></li>
				<li><a href="/article/2">Article 2</a></li>
			</ul>
		`),
		"/article/1": TestArticleHTML("Article 1", "This is the content of article 1."),
		"/article/2": TestArticleHTML("Article 2", "This is the content of article 2."),
	}
}
