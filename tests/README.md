# GoCrawl Test Infrastructure

This directory contains the feature and integration test infrastructure for GoCrawl.

## Structure

```
tests/
├── features/          # CLI feature tests (to be created in Phase 2)
├── integration/      # Integration tests (to be created in Phase 3)
├── contracts/        # Contract tests (to be created in Phase 4)
└── helpers/          # Test utilities and helpers
    ├── elasticsearch.go  # Elasticsearch container management
    ├── fixtures.go       # Test data builders
    ├── server.go         # HTTP test server helpers
    └── assertions.go     # Custom test assertions
```

## Setup

### Dependencies

The test infrastructure requires `testcontainers-go` for managing Elasticsearch containers. Add these dependencies:

```bash
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/testcontainers/testcontainers-go/modules/elasticsearch@latest
go mod tidy
```

**Note:** Docker must be running for testcontainers to work.

### Running Tests

```bash
# Run all tests
task test:all

# Run only unit tests (fast, no containers)
task test:unit

# Run integration tests (requires Docker)
task test:integration

# Run feature tests (requires Docker)
task test:features

# Run with coverage
task test:cover
```

## Helpers

### Elasticsearch Container

The `helpers` package provides `StartElasticsearch()` to manage test Elasticsearch instances:

```go
import "github.com/jonesrussell/gocrawl/tests/helpers"

ctx := context.Background()
es, err := helpers.StartElasticsearch(ctx)
defer es.Stop(ctx)

// Use es.Address or es.GetAddresses() in your config
```

### Test Fixtures

Create test data using builder functions:

```go
// Create a test source
source := helpers.TestSource("test-site", "http://example.com",
    helpers.WithMaxDepth(3),
    helpers.WithRateLimit(2*time.Second),
)

// Create test articles
article := helpers.TestArticle("Test Title", "Test content",
    helpers.WithArticleID("article-1"),
    helpers.WithArticleSource("http://example.com/article"),
)

// Create test pages
page := helpers.TestPage("http://example.com/page", "Page Title", "Page content")
```

### Mock HTTP Servers

Create mock websites for crawling tests:

```go
// Simple mock server
content := map[string]string{
    "/": helpers.TestHTMLPage("Home", "Welcome"),
    "/page1": helpers.TestHTMLPage("Page 1", "Content"),
}
server := helpers.StartTestServer(content)
defer server.Close()

// Use server.URL as the crawl target
```

### Test Assertions

Use custom assertions for Elasticsearch:

```go
helpers.AssertIndexExists(t, storage, ctx, "test_index")
helpers.AssertDocumentIndexed(t, storage, ctx, "test_index", "doc-id")
helpers.AssertDocumentCount(t, storage, ctx, "test_index", 10)
```

## Test Categories

### Unit Tests (`test:unit`)
- Fast execution
- No external dependencies
- Use `-short` flag
- Mock external services

### Integration Tests (`test:integration`)
- Test component interactions
- Use real Elasticsearch (via testcontainers)
- Slower execution
- Tag with `Integration` in test name

### Feature Tests (`test:features`)
- Test CLI commands end-to-end
- Full workflow validation
- Use real Elasticsearch and mock HTTP servers

## Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Clean up resources** with `defer` or `t.Cleanup()`
3. **Skip integration tests in short mode**: `if testing.Short() { t.Skip() }`
4. **Use meaningful test names**: `TestCrawl_WithMaxDepthFlag_LimitsDepth`
5. **Keep tests independent** - no shared state
6. **Use test suites** for complex test scenarios

## Example Test Structure

```go
package integration_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/suite"
    "github.com/jonesrussell/gocrawl/tests/helpers"
)

type EndToEndSuite struct {
    suite.Suite
    es         *helpers.ElasticsearchContainer
    mockServer *httptest.Server
}

func (s *EndToEndSuite) SetupSuite() {
    ctx := context.Background()
    es, err := helpers.StartElasticsearch(ctx)
    s.Require().NoError(err)
    s.es = es
    
    content := helpers.CreateArticlePages()
    s.mockServer = helpers.StartTestServer(content)
}

func (s *EndToEndSuite) TearDownSuite() {
    if s.mockServer != nil {
        s.mockServer.Close()
    }
    if s.es != nil {
        ctx := context.Background()
        _ = s.es.Stop(ctx)
    }
}

func (s *EndToEndSuite) TestFullCrawlPipeline() {
    // Test implementation
}

func TestEndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }
    suite.Run(t, new(EndToEndSuite))
}
```

## CI/CD Integration

Integration tests require Docker. Ensure your CI environment:
- Has Docker installed and running
- Has sufficient resources (memory, CPU)
- Allows container creation

See `.github/workflows/test.yml` for GitHub Actions example.
