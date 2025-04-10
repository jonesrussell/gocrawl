# Config Package

The config package provides configuration management for the GoCrawl application. It handles loading, validation, and access to configuration values from both YAML files and environment variables using Viper.

## Components

### Config (`config.go`)
The main configuration implementation that manages application settings.

Key features:
- Loads configuration from YAML files
- Supports environment variable overrides
- Validates configuration values
- Provides type-safe configuration access
- Handles configuration updates
- Manages default values

### Module (`module.go`)
Manages dependency injection for the config package using the fx framework.

Key features:
- Provides fx module configuration
- Manages configuration dependencies
- Integrates with the main application
- Handles constructor injection
- Provides HTTP transport configuration

### Selectors (`selectors.go`)
Defines CSS selectors for extracting content from web pages.

Key features:
- Provides comprehensive article selectors
- Supports metadata extraction
- Includes default selector configurations
- Handles OpenGraph metadata
- Supports JSON-LD structured data
- Manages content exclusion rules

## Configuration Structure

```yaml
app:
  environment: "development"
  name: "gocrawl"
  version: "1.0.0"

crawler:
  base_url: "https://example.com"
  max_depth: 3
  rate_limit: "1s"
  random_delay: "500ms"
  index_name: "articles"
  content_index_name: "content"
  source_file: "sources.yml"
  parallelism: 2

elasticsearch:
  url: "http://localhost:9200"
  username: "elastic"
  password: "password"
  api_key: ""
  index_name: "articles"
  skip_tls: false

log:
  level: "info"
  debug: false
```

## Usage Example

```go
// Initialize configuration with a custom file
cfg, err := config.InitializeConfig("config.yml")
if err != nil {
    log.Fatal(err)
}

// Access configuration values
baseURL := cfg.Crawler.BaseURL
maxDepth := cfg.Crawler.MaxDepth

// Update configuration values
cfg.Crawler.SetMaxDepth(5)
cfg.Crawler.SetRateLimit(2 * time.Second)

// Get default article selectors
selectors := config.DefaultArticleSelectors()
```

## Key Features

### 1. Configuration Management
- YAML file support
- Environment variable binding
- Default value management
- Configuration validation
- Type-safe access
- Dynamic updates

### 2. Dependency Injection
- fx module integration
- Constructor injection
- HTTP transport configuration
- Component lifecycle management
- Error handling

### 3. Content Extraction
- Article selectors
- Metadata extraction
- OpenGraph support
- JSON-LD support
- Default configurations
- Customizable selectors

### 4. Error Handling
- Validation errors
- Loading errors
- Type conversion errors
- Missing value handling
- Default value fallbacks

## Environment Variables

Essential environment variables:
- `APP_ENV`: Application environment
- `LOG_LEVEL`: Logging level
- `APP_DEBUG`: Debug mode flag
- `CRAWLER_BASE_URL`: Starting URL
- `CRAWLER_MAX_DEPTH`: Maximum crawl depth
- `CRAWLER_RATE_LIMIT`: Request rate limit
- `ELASTICSEARCH_URL`: Elasticsearch URL
- `ELASTICSEARCH_USERNAME`: Elasticsearch username
- `ELASTICSEARCH_PASSWORD`: Elasticsearch password
- `ELASTICSEARCH_INDEX_NAME`: Default index name

## Testing

The package includes comprehensive tests:
- Configuration loading tests
- Validation tests
- Environment variable tests
- Selector tests
- Error handling tests
- Integration tests

## Best Practices

1. **Configuration Files**
   - Use YAML for readability
   - Include comments
   - Use consistent naming
   - Validate all values
   - Set reasonable defaults
   - Document all options

2. **Environment Variables**
   - Use clear naming
   - Document requirements
   - Provide examples
   - Handle missing values
   - Use secure defaults
   - Validate inputs

3. **Selectors**
   - Use specific selectors
   - Include fallbacks
   - Document patterns
   - Test thoroughly
   - Handle missing content
   - Support common formats

4. **Error Handling**
   - Provide clear messages
   - Include context
   - Handle all cases
   - Log appropriately
   - Use custom errors
   - Validate early

## Development

When modifying the config package:
1. Update tests for new functionality
2. Document configuration changes
3. Update validation logic
4. Maintain backward compatibility
5. Follow error handling patterns
6. Update logging appropriately
7. Validate all inputs
8. Handle edge cases
9. Keep documentation current 