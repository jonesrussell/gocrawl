#!/bin/bash

# Generate mocks for the GoCrawl project
# Usage: ./scripts/generate_mocks.sh

set -e

# Create mock directories if they don't exist
mkdir -p testutils/mocks/{api,config,crawler,logger,sources,models,storage}

# API mocks
mockgen -source=internal/api/api.go -destination=testutils/mocks/api/api.go -package=api
mockgen -source=internal/api/module.go -destination=testutils/mocks/api/module.go -package=api
mockgen -source=internal/api/indexing.go -destination=testutils/mocks/api/indexing.go -package=api
mockgen -source=internal/interfaces/index_manager.go -destination=testutils/mocks/api/index_manager.go -package=api

# Config mocks
mockgen -source=internal/config/interface.go -destination=testutils/mocks/config/config.go -package=config

# Crawler mocks
mockgen -source=internal/crawler/interfaces.go -destination=testutils/mocks/crawler/crawler.go -package=crawler

# Logger mocks
mockgen -source=internal/logger/logger.go -destination=testutils/mocks/logger/logger.go -package=logger

# Sources mocks
mockgen -source=internal/sources/interface.go -destination=testutils/mocks/sources/sources.go -package=sources

# Models mocks
mockgen -source=internal/models/content.go -destination=testutils/mocks/models/content_processor.go -package=models

# Storage mocks
mockgen -source=internal/storage/types/interface.go -destination=testutils/mocks/storage/storage.go -package=storage

# Format the generated files
go fmt ./testutils/mocks/... 