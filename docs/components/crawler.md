# Crawler Component

The Crawler is the core orchestrator component that manages the web crawling process. It coordinates between the collector, storage, and logger components while handling configuration and error management.

## Interface

```go
type Crawler interface {
    Start(ctx context.Context) error
    Stop() error
    SetMaxDepth(depth int)
    SetRateLimit(duration time.Duration)
}
```

## Dependencies

```go
type Crawler struct {
    Collector *colly.Collector
    Storage   storage.Storage
    Logger    logger.Interface
    Config    *config.Config
}
```

## Configuration

The crawler accepts several configuration parameters:

```go
type CrawlerConfig struct {
    MaxDepth      int           // Maximum crawling depth
    RateLimit     time.Duration // Time between requests
    BaseURL       string        // Starting URL for crawling
    AllowedDomain string        // Restrict crawling to this domain
}
```

## Usage Example

```go
crawler := NewCrawler(CrawlerParams{
    Collector: collector,
    Storage:   storage,
    Logger:    logger,
    Config:    config,
})

err := crawler.Start(context.Background())
if err != nil {
    log.Fatal(err)
}
```

## Key Features

### 1. Crawling Process Management

```go
func (c *Crawler) Start(ctx context.Context) error {
    c.Logger.Info("Starting crawler", "url", c.Config.BaseURL)
    
    // Configure collector
    c.configureCollector()
    
    // Start crawling
    return c.Collector.Visit(c.Config.BaseURL)
}
```

### 2. Content Processing

```go
func (c *Crawler) processPage(content string, url string) error {
    docID := generateDocID(url)
    
    return c.Storage.IndexDocument(ctx, c.Config.IndexName, docID, map[string]interface{}{
        "url":     url,
        "content": content,
        "crawled": time.Now(),
    })
}
```

### 3. Error Handling

```go
func (c *Crawler) handleError(err error, url string) {
    c.Logger.Error("Crawling error",
        "url", url,
        "error", err,
    )
    
    // Implement retry logic if needed
    if isRetryableError(err) {
        c.retryQueue.Add(url)
    }
}
```

## Event Handlers

The crawler sets up various event handlers for the collector:

```go
func (c *Crawler) configureCollector() {
    // Handle successful page visits
    c.Collector.OnHTML("body", func(e *colly.HTMLElement) {
        c.processPage(e.Text, e.Request.URL.String())
    })

    // Handle errors
    c.Collector.OnError(func(r *colly.Response, err error) {
        c.handleError(err, r.Request.URL.String())
    })

    // Handle rate limiting
    c.Collector.OnRequest(func(r *colly.Request) {
        c.Logger.Debug("Visiting", "url", r.URL.String())
    })
}
```

## Rate Limiting

The crawler implements rate limiting to be respectful to target websites:

```go
func (c *Crawler) SetRateLimit(duration time.Duration) {
    c.Collector.SetRequestDelay(duration)
    c.Logger.Info("Rate limit set", "duration", duration)
}
```

## Testing

Example of testing the crawler component:

```go
func TestCrawler_Start(t *testing.T) {
    // Create mock dependencies
    mockCollector := collector.NewMockCollector()
    mockStorage := storage.NewMockStorage()
    mockLogger := logger.NewMockLogger()
    
    crawler := NewCrawler(CrawlerParams{
        Collector: mockCollector,
        Storage:   mockStorage,
        Logger:    mockLogger,
        Config:    testConfig,
    })
    
    // Test crawling
    err := crawler.Start(context.Background())
    assert.NoError(t, err)
    
    // Verify expectations
    mockStorage.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}
```

## Best Practices

1. **Resource Management**
   - Implement proper cleanup in Stop()
   - Use context for cancellation
   - Close all connections properly

2. **Error Handling**
   - Use structured errors
   - Implement retry mechanisms
   - Log errors with context

3. **Configuration**
   - Validate all configuration parameters
   - Use reasonable defaults
   - Document all options

4. **Testing**
   - Use dependency injection for testability
   - Create mock implementations
   - Test error scenarios
