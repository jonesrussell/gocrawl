# Go Web Crawler

A simple web crawler built in Go that fetches and processes web pages, storing results in Elasticsearch. This project demonstrates the use of Cobra for CLI, Colly for web scraping, Elasticsearch for storage, and Zap for logging.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Tasks](#tasks)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- Command-line interface using Cobra
- Crawl web pages with configurable depth and rate limits
- Store crawled data in Elasticsearch
- Log crawling activities using structured logging
- Easily configurable through environment variables and flags

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/jonesrussell/gocrawl.git
   cd gocrawl
   ```

2. **Install dependencies:**

   Make sure you have Go and Elasticsearch installed. Then, run:

   ```bash
   go mod tidy
   ```

3. **Create a `config.yml` file** (optional):

   You can create a `config.yml` file in the root directory to set environment variables for configuration:

   ```yaml
   APP_ENV: development
   LOG_LEVEL: debug
   APP_DEBUG: true

   ELASTICSEARCH_URL: https://localhost:9200
   ELASTICSEARCH_USERNAME: elastic
   ELASTICSEARCH_PASSWORD: your_password
   ELASTICSEARCH_API_KEY: your_api_key
   ELASTICSEARCH_INDEX_NAME: articles
   ELASTICSEARCH_SKIP_TLS: true

   CRAWLER_BASE_URL: https://www.elliotlaketoday.com/opp-beat
   CRAWLER_MAX_DEPTH: 2
   CRAWLER_RATE_LIMIT: 2s

   # Elasticsearch Configuration
   ELASTICSEARCH_HOSTS: http://localhost:9200
   ELASTICSEARCH_INDEX_PREFIX: gocrawl
   ELASTICSEARCH_MAX_RETRIES: 3
   ELASTICSEARCH_RETRY_INITIAL_WAIT: 1s
   ELASTICSEARCH_RETRY_MAX_WAIT: 30s
   ```

## Usage

The crawler provides several commands through its CLI:

```bash
# Show available commands
gocrawl help

# Start crawling
gocrawl crawl https://example.com

# Search crawled content
gocrawl search "search term"

# Get help for crawl command
gocrawl crawl --help
```

### Crawl Command Options

```bash
gocrawl crawl [url] [flags]

Flags:
  -d, --depth int          Maximum crawl depth (default 2)
  -r, --rate-limit string  Rate limit between requests (default "1s")
  -h, --help              Help for crawl command
```

### Search Command Options

```bash
gocrawl search [query] [flags]

Flags:
  -i, --index string       Elasticsearch index to search (default "crawled_pages")
  -s, --size int          Number of results to return (default 10)
  -h, --help              Help for search command
```

## Configuration

### Server Configuration

The HTTP server can be configured in three ways (in order of precedence):

1. Through `config.yaml`:
```yaml
server:
  address: ":3000"  # Change the port to 3000
  read_timeout: 10s
  write_timeout: 30s
  idle_timeout: 60s
```

2. Through environment variable:
```bash
# Set port to 4000
GOCRAWL_PORT=4000 go run main.go httpd

# Or with colon prefix
GOCRAWL_PORT=:4000 go run main.go httpd
```

3. Default port (8080) if no configuration is provided

## Tasks

This project uses a `Taskfile.yml` for task automation. Here are the available tasks:

- **build**: Build the web crawler executable
- **lint**: Lint the Go code
- **test**: Run the tests
- **docs**: Generate CLI documentation

## Testing

To run the tests, use:

```bash
task test
```

To run tests with coverage reporting:

```bash
task test:cover
```

This will:
- Generate a coverage profile in `coverage/coverage.out`
- Create an HTML coverage report in `coverage/coverage.html`
- Display a function-level coverage summary in the terminal

Current test coverage varies across packages:
- High coverage (80%+): metrics, config/logging, config/app, config/server, config/sources
- Medium coverage (40-80%): api/middleware, crawler/events, api, config/elasticsearch
- Low coverage (<40%): content, crawler, storage
- Several packages need test files added

The project uses mock interfaces for testing. To generate or update mocks, run:

```bash
task generate
```

Make sure Elasticsearch is running locally or update the test configuration to point to your Elasticsearch instance.

### CI/CD Pipeline

The project uses GitHub Actions for continuous integration. The workflow:

1. Checks out the code
2. Caches Go modules
3. Sets up Go 1.24
4. Installs dependencies
5. Generates mocks
6. Builds the project
7. Runs tests
8. Runs benchmarks

The workflow runs on pull requests to the main branch and uses test-specific environment variables for configuration.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
