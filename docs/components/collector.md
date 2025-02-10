# Collector Component

The Collector component, built on top of the Colly framework, handles the actual web page collection process.

## Interface

```go
type Collector interface {
    Visit(url string) error
    SetMaxDepth(depth int)
    SetRequestDelay(delay time.Duration)
    OnHTML(selector string, callback HTMLCallback)
    OnRequest(callback RequestCallback)
    OnError(callback ErrorCallback)
}
```

## Configuration

```go
type CollectorParams struct {
    MaxDepth      int
    RateLimit     time.Duration
    AllowedDomain string
    UserAgent     string
    Debugger      DebuggerInterface
}
```

## Key Features

### 1. Request Configuration

```go
func NewCollector(params Params) (*Collector, error) {
    c := colly.NewCollector(
        colly.MaxDepth(params.MaxDepth),
        colly.AllowedDomains(params.AllowedDomain),
        colly.UserAgent(params.UserAgent),
        colly.Async(true),
    )
    
    // Configure rate limiting
    c.SetRequestDelay(params.RateLimit)
    
    return &Collector{collector: c}, nil
}
```

### 2. URL Filtering

```go
func (c *Collector) configureFilters() {
    c.collector.URLFilters = []*regexp.Regexp{
        regexp.MustCompile(`https?://[^/]+/.*`),
    }
}
```

### 3. Robots.txt Compliance

```go
func (c *Collector) enableRobotsTxt() {
    c.collector.WithTransport(&http.Transport{
        DisableCompression: true,
    })
    c.collector.RobotsTxt = true
}
```

## Best Practices

1. **Rate Limiting**
   - Implement respectful crawling
   - Follow robots.txt rules
   - Use appropriate delays

2. **Error Handling**
   - Handle network errors
   - Implement retries
   - Log failures appropriately

3. **Resource Management**
   - Limit concurrent requests
   - Manage memory usage
   - Clean up resources
