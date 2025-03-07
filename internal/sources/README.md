# Sources Package

The sources package manages the configuration and lifecycle of web content sources for GoCrawl. It handles source configuration loading, validation, and crawling operations through a YAML-based configuration system.

## Components

### Sources (`sources.go`)
The main sources implementation that manages source configurations and crawling operations.

Key features:
- Loads and validates source configurations from YAML
- Manages source lifecycle and crawling operations
- Handles source-specific settings and selectors
- Coordinates with crawler and index manager
- Provides source lookup and management
- Handles graceful shutdown and cleanup

### Module (`module.go`)
Manages dependency injection for the sources package.

Key features:
- Provides fx module configuration
- Manages source dependencies
- Integrates with the main application
- Handles constructor injection
- Provides source configuration grouping

## Interfaces

```go
// IndexManager defines the interface for index management
type IndexManager interface {
    // EnsureIndex ensures that an index exists
    EnsureIndex(ctx context.Context, indexName string) error
}

// CrawlerInterface defines the interface for crawler operations
type CrawlerInterface interface {
    // Start begins crawling from the specified URL
    Start(ctx context.Context, url string) error
    // Stop stops the crawler
    Stop()
}
```

## Configuration

Sources are configured through a YAML file with the following structure:

```yaml
sources:
  - name: "example_source"
    url: "https://example.com"
    article_index: "articles"
    index: "content"
    rate_limit: "1s"
    max_depth: 3
    time:
      - "published_time"
      - "time_ago"
    selectors:
      article:
        container: "article"
        title: "h1"
        body: "article"
        intro: ".intro"
        byline: ".byline"
        published_time: "time"
        time_ago: ".time-ago"
        json_ld: "script[type='application/ld+json']"
        section: ".section"
        keywords: "meta[name='keywords']"
        description: "meta[name='description']"
        og_title: "meta[property='og:title']"
        og_description: "meta[property='og:description']"
        og_image: "meta[property='og:image']"
        og_url: "meta[property='og:url']"
        canonical: "link[rel='canonical']"
        word_count: ".word-count"
        publish_date: ".publish-date"
        category: ".category"
        tags: ".tags"
        author: ".author"
        byline_name: ".byline-name"
```

## Usage Example

```go
// Load sources from configuration file
sources, err := sources.Load("sources.yml")
if err != nil {
    log.Fatal(err)
}

// Set up dependencies
sources.SetCrawler(crawler)
sources.SetIndexManager(indexManager)

// Start crawling a specific source
ctx := context.Background()
err = sources.Start(ctx, "example_source")
if err != nil {
    log.Fatal(err)
}

// Stop crawling when done
sources.Stop()
```

## Key Features

### 1. Source Management
- Loads source configurations from YAML
- Validates source settings
- Manages source lifecycle
- Handles source dependencies
- Provides source lookup
- Manages source state

### 2. Configuration
- YAML-based configuration
- Flexible selector system
- Rate limiting support
- Depth control
- Time field mapping
- Index configuration

### 3. Crawling Control
- Starts and stops crawling
- Handles context cancellation
- Manages crawler lifecycle
- Coordinates with index manager
- Provides graceful shutdown
- Handles error conditions

### 4. Error Handling
- Validates configuration
- Handles file operations
- Manages crawler errors
- Provides context cancellation
- Handles missing sources
- Validates required fields

## Testing

The package includes comprehensive tests:
- Unit tests for source management
- Configuration validation tests
- YAML parsing tests
- Error handling tests
- Integration tests
- Mock implementations

## Best Practices

1. **Configuration**
   - Use descriptive source names
   - Validate all required fields
   - Use appropriate rate limits
   - Set reasonable depth limits
   - Document selector patterns
   - Use consistent naming

2. **Error Handling**
   - Validate configuration files
   - Handle missing sources
   - Manage crawler errors
   - Handle context cancellation
   - Clean up resources
   - Log error details

3. **Resource Management**
   - Clean up crawler resources
   - Handle context timeouts
   - Manage goroutines
   - Handle file operations
   - Clean up on shutdown
   - Manage memory usage

4. **Logging**
   - Log source operations
   - Include source context
   - Log configuration issues
   - Track crawler status
   - Log error details
   - Monitor performance

## Development

When modifying the sources package:
1. Update tests for new functionality
2. Document configuration changes
3. Update validation logic
4. Maintain backward compatibility
5. Follow error handling patterns
6. Update logging appropriately
7. Validate all inputs
8. Handle edge cases
9. Maintain comprehensive code documentation
   - Document all interfaces with purpose and usage
   - Document all struct fields with descriptions
   - Document all methods with parameters and return values
   - Include examples for complex operations
   - Document error handling and edge cases
   - Keep documentation up to date with code changes 