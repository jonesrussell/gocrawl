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

### 6. Migration Plan
1. Create new packages for moved functionality
2. Update interfaces in common package
3. Implement event handlers for content discovery
4. Update module dependencies
5. Update tests
6. Update documentation 