# Indexing Package TODO

## Overview
New package to handle all index-related operations, separating these concerns from crawler and storage packages.

## Core Responsibilities
1. Index management
   - Creation
   - Updates
   - Deletion
   - Mapping management
2. Document operations
   - Indexing
   - Updating
   - Deletion
3. Search operations
   - Basic search
   - Aggregations
   - Filtering

## Proposed Structure
```
indexing/
├── module.go         # DI setup
├── index.go         # Index management
├── document.go      # Document operations
├── search.go        # Search operations
├── mapping.go       # Index mapping definitions
└── options.go       # Configuration options
```

## Interface Definitions

### 1. Index Management
```go
type IndexManager interface {
    EnsureIndex(ctx context.Context, name string, mapping interface{}) error
    DeleteIndex(ctx context.Context, name string) error
    IndexExists(ctx context.Context, name string) (bool, error)
    UpdateMapping(ctx context.Context, name string, mapping interface{}) error
}
```

### 2. Document Operations
```go
type DocumentManager interface {
    Index(ctx context.Context, index string, id string, doc interface{}) error
    Update(ctx context.Context, index string, id string, doc interface{}) error
    Delete(ctx context.Context, index string, id string) error
    Get(ctx context.Context, index string, id string) (interface{}, error)
}
```

### 3. Search Operations
```go
type SearchManager interface {
    Search(ctx context.Context, index string, query interface{}) ([]interface{}, error)
    Count(ctx context.Context, index string, query interface{}) (int64, error)
    Aggregate(ctx context.Context, index string, aggs interface{}) (interface{}, error)
}
```

## Implementation Tasks
1. [ ] Create package structure
2. [ ] Define interfaces in common package
3. [ ] Implement Elasticsearch-based index manager
4. [ ] Add proper error handling and context support
5. [ ] Implement retry mechanisms
6. [ ] Add metrics and monitoring
7. [ ] Write comprehensive tests
8. [ ] Add documentation

## Migration Plan
1. Create new package
2. Move index-related code from storage package
3. Update crawler to use new interfaces
4. Update dependency injection
5. Add metrics and monitoring
6. Update tests
7. Update documentation

## Dependencies
- github.com/elastic/go-elasticsearch/v8
- go.uber.org/fx for DI
- go.uber.org/zap for logging

## Notes
- All operations should support context for cancellation
- Use proper error wrapping
- Include metrics for monitoring
- Support bulk operations where appropriate
- Include proper logging
- Support configuration via environment variables 