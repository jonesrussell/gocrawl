#!/bin/bash

set -e  # Exit on error

# Create mock directories
mkdir -p testutils/mocks/{api,config,crawler,logger,sources,storage}

# Generate mocks for API interfaces
echo "Generating API mocks..."
mockgen -source=internal/api/api.go -destination=testutils/mocks/api/api.go -package=api
mockgen -source=internal/api/indexing.go -destination=testutils/mocks/api/indexing.go -package=api

# Generate mocks for Config interface
echo "Generating Config mocks..."
mockgen -source=internal/config/config.go -destination=testutils/mocks/config/config.go -package=config

# Generate mocks for Crawler interfaces
echo "Generating Crawler mocks..."
mockgen -source=internal/crawler/crawler.go -destination=testutils/mocks/crawler/crawler.go -package=crawler

# Generate mocks for Logger interface
echo "Generating Logger mocks..."
mockgen -source=internal/logger/logger.go -destination=testutils/mocks/logger/logger.go -package=logger

# Generate mocks for Sources interface
echo "Generating Sources mocks..."
mockgen -source=internal/sources/sources.go -destination=testutils/mocks/sources/sources.go -package=sources

# Generate mocks for Storage interfaces
echo "Generating Storage mocks..."
mockgen -source=internal/storage/types/interface.go -destination=testutils/mocks/storage/storage.go -package=storage

# Format the generated code
echo "Formatting generated code..."
go fmt ./testutils/mocks/...

echo "Mock generation completed successfully!" 