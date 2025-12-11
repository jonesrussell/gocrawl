# Sources API Migration Guide

This document explains how to migrate from the `sources.yml` file-based configuration to the REST API-based configuration using the gosources service.

## Overview

The sources configuration system has been updated to support loading sources from the gosources REST API service instead of reading from a YAML file. This provides several benefits:

- **Centralized Management**: Sources are stored in a database and can be managed via REST API
- **Dynamic Updates**: Sources can be added, updated, or removed without restarting the crawler
- **Better Scalability**: Multiple crawler instances can share the same source configurations
- **Version Control**: The gosources service can track changes to source configurations

## Configuration

### API-Based Configuration (Recommended)

To use the API-based configuration, add the `sources_api_url` setting to your crawler configuration in `config.yaml`:

```yaml
crawler:
  sources_api_url: "http://localhost:8050/api/v1/sources"
```

When `sources_api_url` is set, the crawler will:
1. Connect to the gosources API at the specified URL
2. Fetch all enabled sources from the API
3. Use those sources for crawling

### File-Based Configuration (Deprecated, but still supported)

For backward compatibility, you can still use the file-based configuration by setting `source_file`:

```yaml
crawler:
  source_file: "sources.yml"
```

**Note**: If both `sources_api_url` and `source_file` are set, `sources_api_url` takes precedence.

## API Endpoints

The gosources API provides the following endpoints:

### List All Sources
```bash
GET http://localhost:8050/api/v1/sources
```

Returns:
```json
{
  "sources": [...],
  "count": 10
}
```

### Get a Source by ID
```bash
GET http://localhost:8050/api/v1/sources/{id}
```

### Create a New Source
```bash
POST http://localhost:8050/api/v1/sources
Content-Type: application/json

{
  "name": "Example Source",
  "url": "https://example.com",
  "article_index": "example_articles",
  "page_index": "example_pages",
  "rate_limit": "1s",
  "max_depth": 2,
  "enabled": true,
  "selectors": {
    "article": {...},
    "list": {...},
    "page": {...}
  }
}
```

### Update a Source
```bash
PUT http://localhost:8050/api/v1/sources/{id}
Content-Type: application/json

{
  "name": "Updated Source",
  ...
}
```

### Delete a Source
```bash
DELETE http://localhost:8050/api/v1/sources/{id}
```

## Migration Steps

### Step 1: Ensure gosources Service is Running

Make sure the gosources service is running and accessible at the configured URL (default: `http://localhost:8050`).

### Step 2: Migrate Existing Sources

If you have existing sources in `sources.yml`, you'll need to import them into the gosources database using the API:

```bash
# For each source in sources.yml, create it via the API
curl -X POST http://localhost:8050/api/v1/sources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Example News Site",
    "url": "https://www.example-news.com/news/",
    "article_index": "example_news_articles",
    "page_index": "example_news_content",
    "rate_limit": "1s",
    "max_depth": 2,
    "enabled": true,
    "selectors": {...}
  }'
```

### Step 3: Update Configuration

Update your `config.yaml` to use the API:

```yaml
crawler:
  # Comment out or remove the old source_file setting
  # source_file: "sources.yml"

  # Add the new sources_api_url setting
  sources_api_url: "http://localhost:8050/api/v1/sources"
```

### Step 4: Test the Migration

Run the crawler with the new configuration:

```bash
# List sources to verify the API connection
./gocrawl sources list

# Run a crawl to test
./gocrawl crawl --source "Example News Site"
```

### Step 5: Verify

Check the logs to confirm that sources are being loaded from the API:

```
INFO Loading sources from API url=http://localhost:8050/api/v1/sources
```

## Troubleshooting

### Error: "failed to load sources from API"

**Possible causes:**
- The gosources service is not running
- The API URL is incorrect
- Network connectivity issues

**Solution:**
1. Verify the gosources service is running: `curl http://localhost:8050/api/v1/sources`
2. Check the API URL in your configuration
3. Check firewall/network settings

### Error: "no sources found from API"

**Possible causes:**
- No sources have been created in the gosources database
- All sources are disabled

**Solution:**
1. Verify sources exist: `curl http://localhost:8050/api/v1/sources`
2. Create sources via the API if none exist
3. Ensure sources have `enabled: true`

### Fallback to File-Based Configuration

If you encounter issues with the API and need to temporarily fall back to file-based configuration:

1. Comment out `sources_api_url` in your config.yaml
2. Ensure `source_file` points to a valid sources.yml file
3. Restart the crawler

```yaml
crawler:
  # sources_api_url: "http://localhost:8050/api/v1/sources"
  source_file: "sources.yml"
```

## Source JSON Structure

When creating or updating sources via the API, use this JSON structure:

```json
{
  "name": "string (required)",
  "url": "string (required)",
  "article_index": "string (required)",
  "page_index": "string (required)",
  "rate_limit": "string (e.g., '1s', default: '1s')",
  "max_depth": "integer (default: 2)",
  "time": ["string"],
  "enabled": "boolean (default: true)",
  "city_name": "string (optional)",
  "group_id": "string (optional, UUID)",
  "selectors": {
    "article": {
      "container": "string",
      "title": "string",
      "body": "string",
      "intro": "string",
      "link": "string",
      "image": "string",
      "byline": "string",
      "published_time": "string",
      "time_ago": "string",
      "section": "string",
      "category": "string",
      "article_id": "string",
      "json_ld": "string",
      "keywords": "string",
      "description": "string",
      "og_title": "string",
      "og_description": "string",
      "og_image": "string",
      "og_url": "string",
      "og_type": "string",
      "og_site_name": "string",
      "canonical": "string",
      "author": "string",
      "exclude": ["string"]
    },
    "list": {
      "container": "string",
      "article_cards": "string",
      "article_list": "string",
      "exclude_from_list": ["string"]
    },
    "page": {
      "container": "string",
      "title": "string",
      "content": "string",
      "description": "string",
      "keywords": "string",
      "og_title": "string",
      "og_description": "string",
      "og_image": "string",
      "og_url": "string",
      "canonical": "string",
      "exclude": ["string"]
    }
  }
}
```

## Benefits of API-Based Configuration

1. **No File I/O**: Sources are loaded via HTTP, eliminating file system dependencies
2. **Centralized**: All crawler instances can share the same source configurations
3. **Dynamic**: Sources can be added/updated without restarting the crawler
4. **Auditable**: The gosources service can log all changes to sources
5. **Scalable**: Multiple crawlers can run in parallel with synchronized configurations

## Best Practices

1. **Use Environment Variables**: Store the API URL in environment variables for different environments:
   ```yaml
   crawler:
     sources_api_url: "${SOURCES_API_URL}"
   ```

2. **Monitor API Health**: Ensure the gosources service has proper health checks and monitoring

3. **Backup Sources**: Regularly backup the gosources database to prevent data loss

4. **Version Control**: Keep the `sources.yml` file as a backup until you're confident in the API-based setup

5. **Test First**: Test the API configuration in a development environment before deploying to production
