#!/bin/bash

set -e  # Exit on error

# Ensure mockgen is installed (use go.uber.org/mock/mockgen, not github.com/golang/mock/mockgen)
if ! command -v mockgen &> /dev/null; then
    echo "Installing mockgen..."
    go install go.uber.org/mock/mockgen@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Create mock directories
mkdir -p testutils/mocks/{api,config,crawler,logger,sources,storage}

# Generate mocks for API interfaces
echo "Generating API mocks..."
mockgen -source=internal/api/api.go -destination=testutils/mocks/api/api.go -package=api || {
    echo "Failed to generate API mocks" >&2
    exit 1
}
mockgen -source=internal/api/indexing.go -destination=testutils/mocks/api/indexing.go -package=api || {
    echo "Failed to generate API indexing mocks" >&2
    exit 1
}

# Generate mocks for Config interface
echo "Generating Config mocks..."
mockgen -source=internal/config/config.go -destination=testutils/mocks/config/config.go -package=config || {
    echo "Failed to generate Config mocks" >&2
    exit 1
}

# Generate mocks for Crawler interfaces
echo "Generating Crawler mocks..."
mockgen -source=internal/crawler/crawler.go -destination=testutils/mocks/crawler/crawler.go -package=crawler || {
    echo "Failed to generate Crawler mocks" >&2
    exit 1
}

# Generate mocks for EventHandler interface
echo "Generating EventHandler mocks..."
mockgen -source=internal/crawler/events/eventhandler.go -destination=testutils/mocks/crawler/eventhandler.go -package=crawler || {
    echo "Failed to generate EventHandler mocks" >&2
    exit 1
}

# Generate mocks for Logger interface
echo "Generating Logger mocks..."
mockgen -source=internal/logger/logger.go -destination=testutils/mocks/logger/logger.go -package=logger || {
    echo "Failed to generate Logger mocks" >&2
    exit 1
}

# Generate mocks for Sources interface
echo "Generating Sources mocks..."
mockgen -source=internal/sources/sources.go -destination=testutils/mocks/sources/sources.go -package=sources || {
    echo "Failed to generate Sources mocks" >&2
    exit 1
}

# Generate mocks for Storage interfaces
echo "Generating Storage mocks..."
mockgen -source=internal/storage/types/interface.go -destination=testutils/mocks/storage/storage.go -package=storage || {
    echo "Failed to generate Storage mocks" >&2
    exit 1
}

# Format the generated code
echo "Formatting generated code..."
go fmt ./testutils/mocks/... || {
    echo "Failed to format generated code" >&2
    exit 1
}

echo "Mock generation completed successfully!"
