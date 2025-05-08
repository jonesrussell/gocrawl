#!/bin/bash

# Create mock directories
mkdir -p testutils/mocks/{api,config,crawler,logger,sources,models,storage}

# Generate mocks
mockgen -source=internal/api/api.go -destination=testutils/mocks/api/api.go -package=api
mockgen -source=internal/api/module.go -destination=testutils/mocks/api/module.go -package=api
mockgen -source=internal/api/indexing.go -destination=testutils/mocks/api/indexing.go -package=api
mockgen -source=internal/storage/types/index_manager.go -destination=testutils/mocks/api/index_manager.go -package=api

mockgen -source=internal/config/interface.go -destination=testutils/mocks/config/config.go -package=config

mockgen -source=internal/crawler/interfaces.go -destination=testutils/mocks/crawler/crawler.go -package=crawler

mockgen -source=internal/logger/logger.go -destination=testutils/mocks/logger/logger.go -package=logger

mockgen -source=internal/sources/interface.go -destination=testutils/mocks/sources/sources.go -package=sources

mockgen -source=internal/models/content.go -destination=testutils/mocks/models/content_processor.go -package=models

mockgen -source=internal/storage/types/interface.go -destination=testutils/mocks/storage/storage.go -package=storage

# Format the generated code
go fmt ./testutils/mocks/... 