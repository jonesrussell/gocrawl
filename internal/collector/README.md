# Collector Package

The collector package provides the web page collection functionality for GoCrawl. It manages the actual crawling process using the Colly web scraping framework, handling URL processing, rate limiting, and content extraction.

## Components

### Collector (`collector.go`)
The main collector implementation that orchestrates the web page collection process.

Key features:
- Manages collector lifecycle
- Handles URL validation and processing
- Configures rate limiting and parallelism
- Coordinates with handlers and processors

### Config (`config.go`)
Handles collector configuration and validation.

Key features:
- Manages collector parameters
- Validates configuration values
- Handles source-specific settings
- Provides configuration defaults

### Handlers (`handlers.go`)
Manages event handlers for the collector.

Key features:
- Handles page processing events
- Manages error handling
- Coordinates with processors
- Handles completion signaling

### Content (`content.go`)
Handles content extraction and processing.

Key features:
- Extracts article content
- Processes page content
- Handles content validation
- Manages content storage

### Setup (`setup.go`)
Handles collector setup and configuration.

Key features:
- Creates base collector
- Configures collector settings
- Sets up rate limiting
- Handles domain restrictions

## Interfaces

```go
// Logger defines the interface for logging operations
type Logger interface {
    Debug(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Info(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
}

// ArticleProcessor defines the interface for processing articles
type ArticleProcessor interface {
    Process(article interface{}) error
}

// ContentProcessor defines the interface for processing content
type ContentProcessor interface {
    Process(content string) (string, error)
}
```

## Configuration

The collector accepts configuration through the `Config` struct:

```go
type Config struct {
    // BaseURL is the starting URL for crawling
    BaseURL string
    // MaxDepth is the maximum crawling depth
    MaxDepth int
    // RateLimit is the time between requests
    RateLimit string
    // Parallelism is the number of concurrent requests
    Parallelism int
    // RandomDelay adds random delay between requests
    RandomDelay time.Duration
    // Debugger handles debugging operations
    Debugger *logger.CollyDebugger
    // Logger provides logging capabilities
    Logger logger.Interface
    // Source contains source-specific configuration
    Source config.Source
    // ArticleProcessor handles article processing
    ArticleProcessor models.ContentProcessor
    // ContentProcessor handles content processing
    ContentProcessor models.ContentProcessor
}
```

## Usage Example

```go
// Create collector parameters
params := collector.Params{
    ArticleProcessor: articleProcessor,
    ContentProcessor: contentProcessor,
    BaseURL:         "https://example.com",
    Context:         ctx,
    Debugger:        debugger,
    Logger:          logger,
    MaxDepth:        3,
    Parallelism:     2,
    RandomDelay:     time.Second,
    RateLimit:       time.Second,
    Source:          sourceConfig,
}

// Create new collector
result, err := collector.New(params)
if err != nil {
    log.Fatal(err)
}

// Use the collector
err = result.Collector.Visit("https://example.com")
if err != nil {
    log.Fatal(err)
}

// Wait for completion
<-result.Done
```

## Key Features

### 1. URL Processing
- Validates URLs
- Handles domain restrictions
- Manages crawling depth
- Processes relative URLs

### 2. Rate Limiting
- Implements request delays
- Handles concurrent requests
- Manages domain-specific limits
- Provides random delays

### 3. Content Extraction
- Extracts article content
- Processes page content
- Handles content validation
- Manages content storage

### 4. Error Handling
- Provides detailed error messages
- Implements graceful recovery
- Handles connection errors
- Manages resource cleanup

## Testing

The package includes comprehensive tests:
- Unit tests for each component
- Integration tests for the collector
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

When modifying the collector:
1. Update tests for new functionality
2. Document interface changes
3. Update configuration handling
4. Maintain backward compatibility
5. Follow error handling patterns 