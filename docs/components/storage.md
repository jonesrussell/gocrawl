# Storage Component

The Storage component handles data persistence using Elasticsearch as the backend storage system.

## Interface

```go
type Storage interface {
    // IndexDocument stores a document in the specified index
    IndexDocument(ctx context.Context, index string, docID string, document interface{}) error
    
    // TestConnection verifies the connection to the storage backend
    TestConnection(ctx context.Context) error
}
```

## Configuration

```go
type StorageConfig struct {
    ElasticURL      string
    ElasticPassword string
    ElasticAPIKey   string
    IndexName       string
}
```

## Key Features

### 1. Document Indexing

```go
func (s *ElasticsearchStorage) IndexDocument(ctx context.Context, index, docID string, document interface{}) error {
    data, err := json.Marshal(document)
    if err != nil {
        return fmt.Errorf("error marshaling document: %w", err)
    }
    
    res, err := s.ESClient.Index(
        index,
        bytes.NewReader(data),
        s.ESClient.Index.WithDocumentID(docID),
        s.ESClient.Index.WithContext(ctx),
    )
    if err != nil {
        return fmt.Errorf("error indexing document: %w", err)
    }
    defer res.Body.Close()

    if res.IsError() {
        return fmt.Errorf("error indexing document: %s", res.String())
    }

    return nil
}
```

### 2. Connection Management

```go
func NewStorage(config *config.Config) (*ElasticsearchStorage, error) {
    cfg := elasticsearch.Config{
        Addresses: []string{config.ElasticURL},
        APIKey:    config.ElasticAPIKey,
        Username:  "elastic",
        Password:  config.ElasticPassword,
    }
    
    client, err := elasticsearch.NewClient(cfg)
    if err != nil {
        return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
    }
    
    return &ElasticsearchStorage{ESClient: client}, nil
}
```

## Best Practices

1. **Error Handling**
   - Implement retries for transient failures.
   - Use context for timeouts.
   - Log detailed error information.

2. **Performance**
   - Use bulk operations where possible.
   - Implement connection pooling.
   - Monitor resource usage.
