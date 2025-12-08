# GoCrawl

A web crawler and search engine built with Go. It crawls websites, extracts content, and stores it in Elasticsearch for efficient searching.

## Features

- Web crawling with configurable rules
- Content extraction and processing
- Elasticsearch storage and search
- REST API for searching content
- Job scheduling for automated crawling

## Prerequisites

- Go 1.24 or later
- Elasticsearch 8.x
- Docker (optional)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/jonesrussell/gocrawl.git
cd gocrawl
```

2. Install dependencies:
```bash
go mod download
```

3. Configure Elasticsearch:
```bash
cp .env.example .env
# Edit .env with your Elasticsearch settings
```

4. Build and run:
```bash
go build -o bin/gocrawl
./bin/gocrawl
```

## Basic Usage

List available sources:
```bash
./bin/gocrawl sources list
```

Start crawling a source:
```bash
./bin/gocrawl crawl <source-name>
```

Search content:
```bash
./bin/gocrawl search "your search query"
```

## Development

Run tests:
```bash
go test ./...
```

Run linter:
```bash
task lint
```

## License

MIT
