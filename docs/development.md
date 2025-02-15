# Development Guide

## Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- Task (task runner)
- Visual Studio Code with Remote Containers extension (recommended)

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/jonesrussell/gocrawl.git
cd gocrawl
```

### 2. Environment Configuration

Create a `.env` file in the project root:

```bash
# Application
APP_NAME=gocrawl
APP_ENV=development
APP_DEBUG=true
LOG_LEVEL=debug

# Elasticsearch
ELASTIC_URL=http://elasticsearch:9200
ELASTIC_PASSWORD=your_password
ELASTIC_INDEX_NAME=articles
```

### 3. Development Container

The project includes a development container configuration:

1. Open VS Code.
2. Install the "Remote - Containers" extension.
3. Open the project folder.
4. Click "Reopen in Container" when prompted.
5. Wait for the container to build and initialize.

## Development Workflow

### Running Tasks

The project uses `Taskfile.yml` for common operations:

```bash
# Run the crawler
task crawl

# Build the application
task build

# Run tests
task test

# Run linters
task lint
```

### Code Structure

```
.
├── cmd/
│   └── gocrawl/
│       └── main.go
├── internal/
│   ├── collector/
│   ├── config/
│   ├── crawler/
│   ├── logger/
│   └── storage/
├── docs/
├── .devcontainer/
├── Taskfile.yml
└── go.mod
```

## Testing

### Running Tests

```bash
# Run all tests
task test

# Run tests with coverage
go test -v -cover ./...

# Run tests for a specific package
go test -v ./internal/crawler
```

### Writing Tests

- Place tests in `*_test.go` files.
- Use table-driven tests where appropriate.
- Utilize mock implementations for external dependencies.
- Aim for >80% test coverage.

Example test:

```go
func TestCrawler(t *testing.T) {
    // Setup test cases
    tests := []struct {
        name     string
        url      string
        maxDepth int
        want     error
    }{
        {
            name:     "valid url",
            url:      "https://example.com",
            maxDepth: 2,
            want:     nil,
        },
    }

    // Run test cases
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Linting

The project uses `golangci-lint` with custom configuration:

```bash
# Run linters
task lint

# Fix auto-fixable issues
golangci-lint run --fix
```

## Debugging

### VS Code Configuration

Launch configuration for debugging (`.vscode/launch.json`):

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Crawler",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/main.go",
            "args": ["-url=https://example.com", "-maxDepth=2"]
        }
    ]
}
```

### Using Delve

```bash
# Start delve
dlv debug main.go

# Set breakpoint
break main.go:42

# Continue execution
continue
```

## Contributing

1. Create a new branch for your feature/fix.
2. Write tests for new functionality.
3. Ensure all tests pass and linters are happy.
4. Submit a pull request.

### Commit Messages

Follow conventional commits format:

```
feat: add new crawler feature
fix: resolve rate limiting issue
docs: update API documentation
test: add integration tests for storage
```

## Troubleshooting

### Common Issues

1. Elasticsearch Connection
```bash
# Check Elasticsearch status
curl -X GET "localhost:9200/_cluster/health"
```

2. Rate Limiting
```bash
# Adjust rate limit for testing
./bin/gocrawl -rateLimit=5s
```

3. Container Issues
```bash
# Rebuild container
docker-compose build --no-cache
```

### Logs

- Development logs are written to stdout.
- Production logs are in JSON format.
- Use appropriate log levels for debugging.
