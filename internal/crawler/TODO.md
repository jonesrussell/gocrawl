# Crawler Package TODO

## Current Issues
1. Too many direct dependencies:
   - Storage interface
   - Article service
   - Index service
   - Content processor
   - Logger
   - Config

2. Mixed responsibilities:
   - Crawling
   - Article processing
   - Index management
   - Content processing

## Proposed Changes

### 1. Core Crawler Responsibilities
- Keep only crawling-related functionality
- Move to event-based architecture for content discovery
- Remove direct storage/processing dependencies

### 2. Interface Changes
```go
type Crawler interface {
    Start(ctx context.Context, url string) error
    Stop()
    OnContentDiscovered(handler func(ctx context.Context, content *models.Content) error)
    SetRateLimit(duration time.Duration)
    SetMaxDepth(depth int)
}

// Parameters for fx dependency injection
type CrawlerParams struct {
    fx.In
    
    Lifecycle fx.Lifecycle
    Logger    logger.Interface
    Config    *config.Config
}

// Result for fx dependency injection
type CrawlerResult struct {
    fx.Out
    
    Crawler Crawler
    Done    chan struct{} `name:"crawlDone"`
}
```

### 3. Dependencies to Move
- Move article processing to `internal/processor`
- Move index management to `internal/indexing`
- Move storage operations to `internal/storage`
- Move content processing to `internal/processor`

### 4. New Structure
```
crawler/
├── crawler.go       # Core crawler implementation
├── module.go        # DI setup
├── options.go       # Crawler options
├── collector.go     # Colly collector wrapper
└── handlers.go      # Event handlers
```

### 5. Implementation Tasks
1. [ ] Create new interfaces for core crawler functionality
2. [ ] Remove article processing logic
3. [ ] Remove index management logic
4. [ ] Implement event-based content discovery
5. [ ] Update dependency injection
6. [ ] Add proper error handling and context support
7. [ ] Update tests for new structure

### 6. CLI Integration
1. [ ] Update module providers for CLI commands
2. [ ] Add lifecycle hooks for graceful shutdown
3. [ ] Implement progress reporting for CLI
4. [ ] Add CLI-specific error handling
5. [ ] Support command-line options
6. [ ] Add signal handling (SIGINT, SIGTERM)

### 7. Migration Plan
1. Create new packages for moved functionality
2. Update interfaces in common package
3. Implement event handlers for content discovery
4. Update module dependencies
5. Update CLI integration
6. Update tests
7. Update documentation

### 8. fx Module Setup
```go
var Module = fx.Module("crawler",
    fx.Provide(
        NewCrawler,
        fx.Annotate(
            NewEventBus,
            fx.ResultTags(`name:"crawlerEvents"`),
        ),
    ),
    fx.Invoke(RegisterLifecycle),
)
```

## Dependencies
- github.com/gocolly/colly/v2
- go.uber.org/fx
- go.uber.org/zap
- github.com/spf13/cobra (via cmd package)

## Notes
- Support graceful shutdown
- Implement proper error handling
- Add progress reporting
- Support rate limiting
- Add metrics collection
- Support custom collectors
- Implement retry mechanisms 