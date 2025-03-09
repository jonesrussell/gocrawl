# GoCrawl CLI Commands

This directory contains the command-line interface (CLI) commands for the GoCrawl application.

## Available Commands

### `crawl`
Crawls a single source defined in `sources.yml`.

```bash
gocrawl crawl [source]
```

The source argument must match a name defined in your `sources.yml` configuration file.

Example:
```bash
gocrawl crawl example-blog
```

### `search`
Searches content in Elasticsearch.

```bash
gocrawl search [query] [flags]
```

Flags:
- `--size`: Number of results to return (default: 10)

### `job`
Manages scheduled crawling jobs.

```bash
gocrawl job [command]
```

Commands:
- `start`: Start the job scheduler
- `stop`: Stop the job scheduler
- `list`: List all scheduled jobs
- `add`: Add a new scheduled job
- `remove`: Remove a scheduled job

### `indices`
Manages Elasticsearch indices.

```bash
gocrawl indices [command]
```

Commands:
- `create`: Create new indices
- `delete`: Delete existing indices
- `list`: List all indices
- `mapping`: Show index mappings

### `sources`
Manages crawl sources configuration.

```bash
gocrawl sources [command]
```

Commands:
- `list`: List all configured sources
- `add`: Add a new source
- `edit`: Edit an existing source
- `remove`: Remove a source

## Global Flags

- `--config`: Path to config file (default: config.yaml)
- `--help`: Show help for a command

## Configuration

The application uses a YAML configuration file (`config.yaml` by default) for settings. Key configuration sections include:

- `app`: Application-level settings (environment, name, version)
- `crawler`: Crawler-specific settings (parallelism, rate limits, etc.)
- `elasticsearch`: Elasticsearch connection settings
- `log`: Logging configuration
- `sources`: Source-specific crawling configurations

## Examples

1. Crawl a specific source:
```bash
gocrawl crawl "Example Blog"
```

2. Search for content:
```bash
gocrawl search "breaking news" --size 20
```

3. List all configured sources:
```bash
gocrawl sources list
```

4. Create new indices:
```bash
gocrawl indices create --source "Example Blog"
```

## Error Handling

The CLI provides clear error messages and exit codes:
- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Connection error

## Development

When adding new commands:
1. Create a new file in the `cmd` directory
2. Define the command using `cobra.Command`
3. Add the command to `root.go`
4. Include proper error handling and logging
5. Add tests in the corresponding `*_test.go` file 