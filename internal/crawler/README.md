# Crawler Package

The crawler package is the core component of GoCrawl that manages the web crawling process. It coordinates between the collector, storage, and logger components while handling configuration and error management.

## Components

### Crawler (`crawler.go`)
The main crawler implementation that orchestrates the crawling process.

Key features:
- Manages the crawling lifecycle
- Coordinates between collector and storage
- Handles error management and logging
- Provides configuration management

### Module (`module.go`)
Handles dependency injection and module configuration using `fx`.

Key features:
- Provides crawler dependencies
- Manages lifecycle hooks
- Configures default values
- Handles dependency validation

### Mock (`mock_crawler.go`)
Provides mock implementations for testing.

## Interface

```go
// Interface defines the methods required for a crawler.
// It provides the core functionality for web crawling operations.
type Interface interface {
    // Start begins the crawling process at the specified URL.
    // It manages the crawling lifecycle and handles errors.
    Start(ctx context.Context, url string) error
    // Stop performs cleanup operations when the crawler is stopped.
    Stop()
    // SetCollector sets the collector for the crawler.
    // This allows for dependency injection and testing.
    SetCollector(collector *colly.Collector)
    // SetService sets the article service for the crawler.
    // This allows for dependency injection and testing.
    SetService(service article.Interface)
    // GetBaseURL returns the base URL from the configuration.
    GetBaseURL() string
    // GetIndexManager returns the index service interface.
    GetIndexManager() storage.IndexServiceInterface
}
```

## Dependencies

The crawler requires several dependencies:
- `storage.Interface`: For storing crawled content
- `logger.Interface`: For logging operations
- `article.Interface`: For article processing
- `config.Config`: For configuration management
- `models.ContentProcessor`: For content processing

## Configuration

The crawler accepts configuration through the `CrawlerConfig` struct:

```go
// CrawlerConfig holds crawler-specific configuration
type CrawlerConfig struct {
    // BaseURL is the starting URL for crawling
    BaseURL string
    // MaxDepth is the maximum crawling depth
    MaxDepth int
    // RateLimit is the time between requests
    RateLimit time.Duration
    // RandomDelay adds random delay between requests
    RandomDelay time.Duration
    // IndexName is the name of the Elasticsearch index
    IndexName string
    // ContentIndexName is the name of the content index
    ContentIndexName string
    // SourceFile is the path to the source configuration file
    SourceFile string
    // Parallelism is the number of concurrent requests
    Parallelism int
}
```

## Usage Example

```go
// Create a new crawler with dependencies
crawler := NewCrawler(CrawlerParams{
    Logger:           logger,
    Storage:          storage,
    Debugger:         debugger,
    Config:           config,
    Source:           sourceName,
    IndexService:     indexService,
    ContentProcessor: processors,
})

// Start crawling
err := crawler.Start(ctx, "https://example.com")
if err != nil {
    log.Fatal(err)
}
```

## Key Features

### 1. Crawling Process Management
- Handles URL validation
- Manages concurrent requests
- Implements rate limiting
- Handles errors gracefully

### 2. Content Processing
- Processes articles and content
- Manages content storage
- Handles content indexing
- Implements content validation

### 3. Error Handling
- Provides detailed error messages
- Implements graceful shutdown
- Handles connection errors
- Manages resource cleanup

### 4. Resource Management
- Manages memory usage
- Handles concurrent connections
- Implements timeouts
- Provides cleanup mechanisms

## Testing

The package includes comprehensive tests:
- Unit tests for each component
- Integration tests for the crawler
- Mock implementations for testing
- Test utilities and helpers

## Best Practices

1. **Error Handling**
   - Use descriptive error messages
   - Implement proper error wrapping
   - Handle context cancellation
   - Clean up resources on errors

2. **Resource Management**
   - Use timeouts for operations
   - Implement proper cleanup
   - Handle concurrent operations
   - Manage memory efficiently

3. **Configuration**
   - Validate configuration values
   - Use sensible defaults
   - Document configuration options
   - Handle missing values gracefully

4. **Logging**
   - Log important operations
   - Include relevant context
   - Use appropriate log levels
   - Mask sensitive information

## Development

When modifying the crawler:
1. Update tests for new functionality
2. Document interface changes
3. Update configuration handling
4. Maintain backward compatibility
5. Follow error handling patterns 