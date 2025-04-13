#!/bin/bash

# Generate mocks for the GoCrawl project
# Usage: ./scripts/generate_mocks.sh

set -e

# Create mock directories if they don't exist
mkdir -p testutils/mocks/{api,config,crawler,indices,models,storage}

# API mocks
mockgen -source=internal/api/module.go -destination=testutils/mocks/api/module.go -package=api
mockgen -source=internal/api/indexing.go -destination=testutils/mocks/api/indexing.go -package=api
mockgen -source=internal/interfaces/search.go -destination=testutils/mocks/api/search.go -package=api

# Config mocks
mockgen -source=internal/config/interface.go -destination=testutils/mocks/config/config.go -package=config

# Crawler mocks
mockgen -source=internal/crawler/interfaces.go -destination=testutils/mocks/crawler/crawler.go -package=crawler

# Indices mocks
mockgen -source=internal/logger/logger.go -destination=testutils/mocks/indices/logger.go -package=logger
mockgen -source=internal/sources/interface.go -destination=testutils/mocks/indices/sources.go -package=sources

# Models mocks
mockgen -source=internal/models/content.go -destination=testutils/mocks/models/content_processor.go -package=models

# Storage mocks
mockgen -source=internal/storage/types/interface.go -destination=testutils/mocks/storage/storage.go -package=storage

# Format the generated files
go fmt ./testutils/mocks/... 