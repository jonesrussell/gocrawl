# Processor Package TODO

## Overview
New package to handle all content processing operations, separating these concerns from crawler and storage packages.

## Core Responsibilities
1. Content Processing
   - HTML parsing
   - Content extraction
   - Metadata extraction
   - Content cleaning
2. Article Processing
   - Title extraction
   - Body extraction
   - Date parsing
   - Author extraction
3. Pipeline Management
   - Processing pipeline configuration
   - Pipeline execution
   - Error handling
4. Content Transformation
   - Format conversion
   - Content normalization
   - Language detection
   - Text analysis

## Proposed Structure
```
processor/
├── module.go           # DI setup
├── processor.go        # Core processor implementation
├── pipeline.go         # Pipeline management
├── article.go         # Article processing
├── content.go         # Content processing
├── extractor/         # Content extractors
│   ├── html.go        # HTML extractor
│   ├── json_ld.go     # JSON-LD extractor
│   └── meta.go        # Meta tags extractor
├── transformer/       # Content transformers
│   ├── cleaner.go     # Content cleaner
│   ├── normalizer.go  # Text normalizer
│   └── language.go    # Language detector
└── options.go         # Configuration options
```

## Interface Definitions

### 1. Content Processor
```go
type ContentProcessor interface {
    Process(ctx context.Context, content *models.RawContent) (*models.ProcessedContent, error)
    Configure(options *ProcessorOptions) error
    AddTransformer(transformer Transformer)
    AddExtractor(extractor Extractor)
}
```

### 2. Pipeline Management
```go
type Pipeline interface {
    Execute(ctx context.Context, content interface{}) (interface{}, error)
    AddStep(step PipelineStep)
    SetErrorHandler(handler ErrorHandler)
}
```

### 3. Content Extraction
```go
type Extractor interface {
    Extract(ctx context.Context, content *models.RawContent) (*models.ExtractedContent, error)
    Supports(content *models.RawContent) bool
}
```

### 4. Content Transformation
```go
type Transformer interface {
    Transform(ctx context.Context, content *models.ProcessedContent) error
    Priority() int
}
```

## Implementation Tasks
1. [ ] Create package structure
2. [ ] Define interfaces in common package
3. [ ] Implement core processor
4. [ ] Implement pipeline management
5. [ ] Add HTML content extractor
6. [ ] Add JSON-LD extractor
7. [ ] Add meta tags extractor
8. [ ] Add content transformers
9. [ ] Add proper error handling
10. [ ] Write comprehensive tests
11. [ ] Add documentation

## Migration Plan
1. Create new package
2. Move processing code from crawler and article packages
3. Implement new interfaces
4. Update dependency injection
5. Add content extractors
6. Add content transformers
7. Update tests
8. Update documentation

## Dependencies
- golang.org/x/net/html for HTML parsing
- github.com/PuerkitoBio/goquery for content extraction
- go.uber.org/fx for DI
- go.uber.org/zap for logging

## Notes
- All operations should support context for cancellation
- Use proper error wrapping
- Support parallel processing where appropriate
- Include proper logging
- Support configuration via environment variables
- Support custom extractors and transformers
- Implement proper content validation
- Support content caching
- Add metrics for monitoring processing performance 