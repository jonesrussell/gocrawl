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

### 4. fx Integration
```go
// Parameters for fx dependency injection
type IndexParams struct {
    fx.In
    
    Lifecycle fx.Lifecycle
    Logger    logger.Interface
    Config    *config.Config
}

// Result for fx dependency injection
type IndexResult struct {
    fx.Out
    
    IndexManager    IndexManager
    DocumentManager DocumentManager
    SearchManager   SearchManager
}

var Module = fx.Module("indexing",
    fx.Provide(
        NewIndexManager,
        NewDocumentManager,
        NewSearchManager,
    ),
    fx.Invoke(RegisterLifecycle),
)
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

## CLI Integration
1. [ ] Add index management commands
   - Create index
   - Delete index
   - List indices
   - Update mappings
2. [ ] Add document management commands
   - Index document
   - Delete document
   - Get document
3. [ ] Add search commands
   - Basic search
   - Aggregations
4. [ ] Add progress reporting
5. [ ] Add proper error handling
6. [ ] Support configuration via flags
7. [ ] Add index validation commands

## Migration Plan
1. Create new package
2. Move index-related code from storage package
3. Update crawler to use new interfaces
4. Update dependency injection
5. Add CLI commands
6. Add metrics and monitoring
7. Update tests
8. Update documentation

## Dependencies
- github.com/elastic/go-elasticsearch/v8
- go.uber.org/fx for DI
- go.uber.org/zap for logging
- github.com/spf13/cobra for CLI

## Notes
- All operations should support context for cancellation
- Use proper error wrapping
- Include metrics for monitoring
- Support bulk operations where appropriate
- Include proper logging
- Support configuration via environment variables
- Add index templates support
- Support index aliases
- Implement index lifecycle management
- Add index backup/restore commands
- Support index reindexing
- Add index optimization commands 