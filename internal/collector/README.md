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
- Validates input parameters and configuration
- Creates and configures the base collector
- Sets up event handlers and completion channel

### Config (`config.go`)
Handles collector configuration and validation.

Key features:
- Manages collector parameters using fx.In for dependency injection
- Validates configuration values
- Handles source-specific settings
- Provides configuration defaults
- Manages rate limit parsing and validation
- Handles debugger and logger configuration
- Validates base URL and crawling parameters

### Handlers (`handlers.go`)
Manages event handlers for the collector.

Key features:
- Handles page processing events
- Manages error handling
- Coordinates with processors
- Handles completion signaling
- Manages request and response tracking
- Provides detailed logging of events
- Handles base URL completion detection

### Content (`content.go`)
Handles content extraction and processing.

Key features:
- Extracts article content using multiple detection methods
- Processes page content with type-specific handlers
- Handles content validation
- Manages content storage
- Provides context management for content processing
- Implements article type detection
- Handles link following and processing
- Manages content-specific logging

### Setup (`setup.go`)
Handles collector setup and configuration.

Key features:
- Creates base collector with domain restrictions
- Configures collector settings
- Sets up rate limiting
- Handles domain restrictions
- Validates URLs and protocols
- Configures content processing
- Sets up debugging and logging

### Logging (`logging.go`)
Manages logging configuration for the collector.

Key features:
- Configures request event logging
- Handles response event logging
- Manages error event logging
- Provides structured logging with context
- Integrates with the main logger interface

### Debugger (`debugger.go`)
Provides debugging capabilities for the collector.

Key features:
- Defines debugger interface
- Handles debug event processing
- Provides initialization support
- Ensures compatibility with Colly's debugger

### Selectors (`selectors.go`)
Defines default CSS selectors for content extraction.

Key features:
- Provides default article selectors
- Defines title and date selectors
- Handles author and category selectors
- Supports fallback selectors when source config is missing

### Module (`module.go`)
Manages dependency injection for the collector package.

Key features:
- Provides fx module configuration
- Manages collector dependencies
- Integrates with the main application
- Handles constructor injection

## Interfaces

```go
// Logger defines the interface for logging operations
type Logger interface {
    // Debug logs a debug message with optional fields
    Debug(msg string, fields ...interface{})
    // Error logs an error message with optional fields
    Error(msg string, fields ...interface{})
    // Info logs an informational message with optional fields
    Info(msg string, fields ...interface{})
    // Warn logs a warning message with optional fields
    Warn(msg string, fields ...interface{})
}

// ArticleProcessor defines the interface for processing articles
type ArticleProcessor interface {
    // Process handles the processing of an article
    Process(article interface{}) error
}

// ContentProcessor defines the interface for processing content
type ContentProcessor interface {
    // Process handles the processing of web content
    Process(content string) (string, error)
}

// DebuggerInterface defines the interface for collector debugging
type DebuggerInterface interface {
    // Init initializes the debugger
    Init() error
    // Event handles a debug event from the collector
    Event(e *debug.Event)
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
- Validates URLs and protocols
- Handles domain restrictions
- Manages crawling depth
- Processes relative URLs
- Implements URL pattern matching
- Handles forbidden domains

### 2. Rate Limiting
- Implements request delays
- Handles concurrent requests
- Manages domain-specific limits
- Provides random delays
- Configures parallelism
- Handles rate limit parsing

### 3. Content Extraction
- Extracts article content using multiple methods
- Processes page content with type-specific handlers
- Handles content validation
- Manages content storage
- Implements article detection
- Handles content type routing

### 4. Error Handling
- Provides detailed error messages
- Implements graceful recovery
- Handles connection errors
- Manages resource cleanup
- Validates configuration
- Handles ignored errors

## Testing

The package includes comprehensive tests:
- Unit tests for each component
- Integration tests for the collector
- Mock implementations for testing
- Test utilities and helpers
- Error case coverage
- Configuration validation tests

## Best Practices

1. **Error Handling**
   - Use descriptive error messages
   - Implement proper error wrapping
   - Handle context cancellation
   - Clean up resources on errors
   - Validate input parameters
   - Handle ignored errors appropriately

2. **Resource Management**
   - Use timeouts for operations
   - Implement proper cleanup
   - Handle concurrent operations
   - Manage memory efficiently
   - Handle connection pooling
   - Manage context lifecycle

3. **Configuration**
   - Validate configuration values
   - Use sensible defaults
   - Document configuration options
   - Handle missing values gracefully
   - Parse rate limits correctly
   - Validate URLs and protocols

4. **Logging**
   - Log important operations
   - Include relevant context
   - Use appropriate log levels
   - Mask sensitive information
   - Provide structured logging
   - Include request/response details

## Development

When modifying the collector:
1. Update tests for new functionality
2. Document interface changes
3. Update configuration handling
4. Maintain backward compatibility
5. Follow error handling patterns
6. Update logging appropriately
7. Validate all inputs
8. Handle edge cases 