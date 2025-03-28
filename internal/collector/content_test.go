package collector_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/jonesrussell/gocrawl/internal/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockLogger implements common.Logger for testing
type MockLogger struct {
	mock.Mock
	DebugMessages []string
	ErrorMessages []string
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.ErrorMessages = append(m.ErrorMessages, msg)
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.DebugMessages = append(m.DebugMessages, msg)
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...any) {
	m.Called(msg, fields)
}

func (m *MockLogger) Printf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...any) {
	m.Called(format, args)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// MockProcessor implements Processor for testing
type MockProcessor struct {
	mock.Mock
	ProcessCalls int
}

func (m *MockProcessor) Process(e *colly.HTMLElement) error {
	m.ProcessCalls++
	args := m.Called(e)
	return args.Error(0)
}

func TestContextManager(t *testing.T) {
	t.Parallel()

	ctx := colly.NewContext()
	cm := collector.NewContextManager(ctx)

	t.Run("setHTMLElement", func(t *testing.T) {
		t.Parallel()
		elem := &colly.HTMLElement{
			Response: &colly.Response{},
		}
		cm.SetHTMLElement(elem)
		got, ok := cm.GetHTMLElement()
		assert.True(t, ok)
		assert.Equal(t, elem, got)
	})

	t.Run("markAsArticle", func(t *testing.T) {
		t.Parallel()
		cm.MarkAsArticle()
		assert.True(t, cm.IsArticle())
	})
}

func TestContentLogger(t *testing.T) {
	t.Parallel()

	mockLogger := &MockLogger{
		DebugMessages: make([]string, 0),
		ErrorMessages: make([]string, 0),
	}
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	p := collector.ContentParams{
		Logger: mockLogger,
	}
	log := collector.NewLogger(p)

	t.Run("debug", func(t *testing.T) {
		t.Parallel()
		log.Debug("test message")
		assert.Contains(t, mockLogger.DebugMessages, "test message")
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()
		log.Error("test error")
		assert.Contains(t, mockLogger.ErrorMessages, "test error")
	})
}

func TestArticleDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "article with metadata",
			html: `
				<html>
					<head>
						<meta property="og:type" content="article">
						<meta property="article:published_time" content="2024-01-01T00:00:00Z">
						<meta property="article:author" content="Test Author">
						<meta property="article:section" content="Technology">
					</head>
					<body>
						<article>
							<h1>Test Article</h1>
							<p>Test content</p>
						</article>
					</body>
				</html>
			`,
			expected: true,
		},
		{
			name: "non-article content",
			html: `
				<html>
					<body>
						<div>
							<h1>Welcome</h1>
							<p>This is not an article</p>
						</div>
					</body>
				</html>
			`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			req := &colly.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test",
				},
				Ctx: colly.NewContext(),
			}
			resp := &colly.Response{
				Request: req,
				Body:    []byte(tt.html),
			}
			e := &colly.HTMLElement{
				Request:  req,
				Response: resp,
				DOM:      doc.Selection,
			}
			assert.Equal(t, tt.expected, collector.IsArticleType(e))
		})
	}
}

func TestContentProcessing(t *testing.T) {
	t.Parallel()

	mockLogger := &MockLogger{
		DebugMessages: make([]string, 0),
		ErrorMessages: make([]string, 0),
	}
	mockProcessor := &MockProcessor{}

	tests := []struct {
		name           string
		html           string
		articleProc    collector.Processor
		contentProc    collector.Processor
		expectedCalls  int
		expectedErrors int
	}{
		{
			name: "process article content",
			html: `
				<html>
					<head>
						<meta property="article:published_time" content="2024-01-01T00:00:00Z">
					</head>
					<body>
						<article>
							<h1>Test Article</h1>
							<p>Test content</p>
						</article>
					</body>
				</html>
			`,
			articleProc:    mockProcessor,
			contentProc:    nil,
			expectedCalls:  1,
			expectedErrors: 0,
		},
		{
			name: "no processors available",
			html: `
				<html>
					<body>
						<div>Test content</div>
					</body>
				</html>
			`,
			articleProc:    nil,
			contentProc:    nil,
			expectedCalls:  0,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			req := &colly.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test",
				},
				Ctx: colly.NewContext(),
			}
			resp := &colly.Response{
				Request: req,
				Body:    []byte(tt.html),
			}
			e := &colly.HTMLElement{
				Request:  req,
				Response: resp,
				DOM:      doc.Selection,
			}
			p := collector.ContentParams{
				Logger:           mockLogger,
				ArticleProcessor: tt.articleProc,
				ContentProcessor: tt.contentProc,
			}
			collector.ProcessContent(e, resp, p, collector.NewContextManager(req.Ctx), collector.NewLogger(p))
			assert.Equal(t, tt.expectedCalls, mockProcessor.ProcessCalls)
			assert.Len(t, mockLogger.ErrorMessages, tt.expectedErrors)
		})
	}
}

func TestLinkFollowing(t *testing.T) {
	t.Parallel()

	ignoredErrors := map[string]bool{
		"test error": true,
	}

	tests := []struct {
		name           string
		link           string
		ignoredErrors  map[string]bool
		expectedError  bool
		expectedVisits int
	}{
		{
			name:           "valid link",
			link:           "https://example.com",
			ignoredErrors:  ignoredErrors,
			expectedError:  false,
			expectedVisits: 1,
		},
		{
			name:           "ignored error",
			link:           "https://example.com",
			ignoredErrors:  map[string]bool{"test error": true},
			expectedError:  false,
			expectedVisits: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a request directly
			req := &colly.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test",
				},
				Ctx: colly.NewContext(),
			}

			// Create a mock response
			resp := &colly.Response{
				Request: req,
			}

			// Create a mock HTML element
			e := &colly.HTMLElement{
				Request:  req,
				Response: resp,
			}

			// Call visitLink
			err := collector.VisitLink(e, tt.link, tt.ignoredErrors)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCanProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		articleProc collector.Processor
		contentProc collector.Processor
		expected    bool
	}{
		{
			name:        "has article processor",
			articleProc: &MockProcessor{},
			contentProc: nil,
			expected:    true,
		},
		{
			name:        "has content processor",
			articleProc: nil,
			contentProc: &MockProcessor{},
			expected:    true,
		},
		{
			name:        "no processors",
			articleProc: nil,
			contentProc: nil,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := collector.ContentParams{
				ArticleProcessor: tt.articleProc,
				ContentProcessor: tt.contentProc,
			}
			assert.Equal(t, tt.expected, collector.CanProcess(p))
		})
	}
}
