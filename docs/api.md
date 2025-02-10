# API Documentation

## Core Interfaces

### Storage Interface
The Storage interface defines methods for storing crawled data:

```go
type Storage interface {
    // IndexDocument stores a document in the specified index
    IndexDocument(ctx context.Context, index string, docID string, document interface{}) error
    
    // TestConnection verifies the connection to the storage backend
    TestConnection(ctx context.Context) error
}
```

### Logger Interface
The Logger interface provides structured logging capabilities:

```go
type Interface interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Debug(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Fatalf(msg string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

## Configuration

### Command Line Flags
The application accepts the following command-line flags:

| Flag | Description | Default |
|------|-------------|---------|
| `-url` | The URL to crawl | `http://example.com` |
| `-maxDepth` | Maximum crawl depth | `2` |
| `-rateLimit` | Rate limit between requests | `5s` |

### Environment Variables
Required environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `ELASTIC_URL` | Elasticsearch server URL | Yes |
| `ELASTIC_PASSWORD` | Elasticsearch password | Yes |
| `ELASTIC_API_KEY` | Elasticsearch API key | No |
| `INDEX_NAME` | Name of the Elasticsearch index | Yes |
| `APP_ENV` | Application environment (development/production) | No |
| `APP_NAME` | Application name | No |
| `APP_DEBUG` | Enable debug mode | No |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | No |

## Usage Examples

### Starting the Crawler

```bash
# Basic usage with defaults
./bin/gocrawl -url="https://example.com"

# Advanced usage with all flags
./bin/gocrawl -url="https://example.com" -maxDepth=3 -rateLimit=2s
```

### Storage Example

```go
// Initialize storage
storage, err := storage.NewStorage(config)
if err != nil {
    log.Fatal(err)
}

// Index a document
err = storage.IndexDocument(ctx, "articles", "doc1", map[string]interface{}{
    "url": "https://example.com/article",
    "content": "Article content...",
})
```

### Logger Example

```go
// Initialize logger
logger, err := logger.NewCustomLogger(logger.Params{
    Level: zapcore.InfoLevel,
    AppEnv: "development",
})

// Log messages
logger.Info("Starting crawler", "url", "https://example.com")
logger.Debug("Processing page", "url", "https://example.com/page")
logger.Error("Failed to fetch page", "url", "https://example.com/error", "error", err)
```
