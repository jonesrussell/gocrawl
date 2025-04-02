# GoCrawl Cleanup Plan

## High Priority

### 1. Code Cleanup and Reorganization
- [x] Consolidate Test Utilities
  - [x] Merge `cmd/common/testutil` and `cmd/common/testutils`
  - [x] Move all test utilities to `internal/testutils`
  - [x] Update imports across codebase
  - [x] Remove duplicate test utilities
  - [x] Add proper documentation
- [x] Consolidate Types
  - [x] Move `Job` and `Item` structs to `internal/models`
  - [x] Move storage interface to `internal/storage`
  - [x] Update all imports to use new type locations
  - [x] Remove unused types
  - [x] Add proper documentation
- [ ] Remove Unused Code
  - [ ] Remove `internal/app` package if unused
  - [ ] Remove `internal/processor` package if unused
  - [ ] Remove duplicate TODO files
  - [ ] Clean up unused imports
  - [ ] Remove deprecated code
- [ ] Standardize Package Structure
  - [ ] Move all interfaces to `interface.go`
  - [ ] Move all types to `types.go`
  - [ ] Move all constants to `constants.go`
  - [ ] Move all errors to `errors.go`
  - [ ] Move all tests to `_test` packages
  - [ ] Add proper package documentation
- [ ] Clean Up Documentation
  - [ ] Consolidate all TODOs into main TODO.md
  - [ ] Add proper package documentation
  - [ ] Add proper function documentation
  - [ ] Add proper type documentation
  - [ ] Add proper interface documentation
  - [ ] Add proper examples

### 2. Dependency Injection Simplification
- [x] Simplify Logger Implementation
  - [x] Move logger interface to `internal/logger`
  - [x] Remove `common/types` logger interface
  - [x] Update all imports to use new logger package
  - [x] Simplify zap integration
  - [x] Remove unnecessary wrapper layers
  - [x] Create simple constructor function
  - [x] Add proper error handling
  - [x] Update tests to use new logger
- [x] Remove Unnecessary Abstractions
  - [x] Remove `common/types` package
  - [x] Move types to appropriate packages
  - [x] Remove type aliases
  - [x] Simplify module structure
  - [x] Remove named dependencies
  - [x] Use constructor injection
  - [x] Keep fx for application-level composition only
- [x] Review and Replace Cobra CLI
  - [x] Evaluate current CLI needs
    - [x] List required commands
    - [x] Document usage patterns
    - [x] Identify complexity points
  - [x] Design simpler CLI using standard flag package
    - [x] Create simple command structure
    - [x] Add proper help messages
  - [x] Implement new CLI
    - [x] Create basic command structure
    - [x] Add command implementations
    - [x] Add proper error handling
  - [x] Update tests to use new command structure
    - [x] Add CLI tests
    - [x] Add integration tests
    - [x] Update existing tests
  - [x] Remove Cobra dependency
    - [x] Clean up imports
    - [x] Update documentation
    - [x] Update build scripts

### 3. Module Reorganization
- [ ] Split common module into specific modules
  - [x] Create separate logger module
    - [x] Move logger interface to internal/logger
    - [x] Create logger module for dependency injection
    - [x] Add logger tests
    - [x] Update job command to use new logger module
  - [x] Create separate config module
    - [x] Move config interface to internal/config
    - [x] Create config module for dependency injection
    - [x] Add config tests
    - [ ] Update commands to use new config module
      - [x] Fix type mismatches between old and new config interfaces
      - [x] Update job command to use new config module
      - [ ] Update crawl command to use new config module
      - [ ] Update httpd command to use new config module
  - [x] Create separate sources module
    - [x] Move sources interface to internal/sources
    - [x] Create sources module for dependency injection
    - [x] Add sources tests
    - [ ] Update commands to use new sources module
  - [ ] Create separate storage module
    - [ ] Move storage interface to internal/storage
    - [ ] Create storage module for dependency injection
    - [ ] Add storage tests
    - [ ] Update commands to use new storage module
  - [ ] Create separate metrics module
    - [ ] Move metrics interface to internal/metrics
    - [ ] Create metrics module for dependency injection
    - [ ] Add metrics tests
    - [ ] Update commands to use new metrics module

### 4. Interface Organization and Naming
- [x] Move interfaces to consuming packages
  - [x] Move `ContentProcessor` from `models` to `collector`
  - [x] Move `Logger` interface from `collector` to `common`
  - [x] Move `ArticleProcessor` to `article` package
- [x] Rename interfaces to avoid stuttering
  - [x] `models.ContentProcessor` → `collector.Processor`
  - [x] `collector.ArticleProcessor` → `article.Processor`
  - [x] `collector.ContentProcessor` → `content.Processor`

### 5. Error Handling
- [x] Standardize error handling
  - [x] Use `errors.New` for simple errors
  - [x] Use `fmt.Errorf` with `%w` for wrapped errors
  - [x] Add proper error context
  - [x] Add proper error types
- [ ] Improve HTTP client error handling
  - [ ] Add proper response body closure
    - [ ] Add defer statements
    - [ ] Add error handling
    - [ ] Add tests
  - [ ] Add proper context handling
    - [ ] Add context timeouts
    - [ ] Add context cancellation
    - [ ] Add tests
  - [ ] Add proper timeout handling
    - [ ] Add timeout configuration
    - [ ] Add timeout errors
    - [ ] Add tests

## Medium Priority

### 6. Testing
- [x] Improve test organization
  - [x] Move tests to separate `_test` packages
  - [x] Add proper test helpers
  - [x] Add proper test fixtures
  - [x] Add proper test cleanup
- [x] Add proper test coverage
  - [x] Add edge case tests
  - [x] Add error condition tests
  - [x] Add concurrent operation tests
  - [x] Add lifecycle event tests
- [x] Improve test dependency injection
  - [x] Create test-specific modules
  - [x] Use mock configurations
  - [x] Properly scope test dependencies
  - [x] Add test validation

### 7. Code Organization
- [x] Split large files
  - [x] Split `collector.go`
  - [x] Split `crawler.go`
  - [x] Split `content.go`
- [x] Improve package organization
  - [x] Move common types to `common`
  - [x] Move common interfaces to `common`
  - [x] Move common constants to `common`
  - [x] Move common utilities to `common`

### 8. Documentation
- [x] Improve documentation
  - [x] Add package documentation
  - [x] Add function documentation
  - [x] Add type documentation
  - [x] Add interface documentation
- [ ] Add proper examples
  - [ ] Add usage examples
  - [ ] Add test examples
  - [ ] Add error handling examples
  - [ ] Add configuration examples

## Low Priority

### 9. Configuration
- [x] Improve configuration
  - [x] Add proper validation
  - [x] Add proper defaults
  - [x] Add environment variable support
  - [x] Add file support
- [x] Add proper configuration types
  - [x] Add configuration structs
  - [x] Add configuration methods
  - [x] Add configuration validation
  - [x] Add configuration defaults

### 10. Logging
- [x] Improve logging
  - [x] Add proper log levels
  - [x] Add proper log fields
  - [x] Add proper log context
  - [x] Add proper log formatting
- [x] Add proper logging configuration
  - [x] Add log output configuration
  - [x] Add log rotation
  - [x] Add log filtering
  - [x] Add log formatting

### 11. Metrics and Monitoring
- [x] Add proper metrics
  - [x] Add counter metrics
  - [x] Add gauge metrics
  - [x] Add histogram metrics
  - [x] Add summary metrics
- [x] Add proper monitoring
  - [x] Add health checks
  - [x] Add readiness checks
  - [x] Add liveness checks
  - [x] Add metrics endpoint

### 12. Security
- [x] Improve security
  - [x] Add proper TLS configuration
  - [x] Add proper authentication
  - [x] Add proper authorization
  - [x] Add proper rate limiting
- [x] Add proper security headers
  - [x] Add CORS headers
  - [x] Add CSP headers
  - [x] Add HSTS headers
  - [x] Add XSS headers

## Progress Tracking

### Completed
- Interface Organization and Naming
  - Moved interfaces to consuming packages
  - Renamed interfaces to avoid stuttering
- Dependency Injection
  - Improved fx module organization
  - Added proper lifecycle hooks
  - Added proper error handling in constructors
  - Added proper dependency validation
  - Fixed dependency conflicts
  - Improved test dependency injection
- Error Handling
  - Standardized error handling across the codebase
  - Added proper error wrapping and context
  - Fixed unchecked error returns
- Testing
  - Improved test organization
  - Added comprehensive test coverage
  - Added test helpers and fixtures
  - Added lifecycle event tests
  - Improved test dependency injection
- Code Organization
  - Split large files into smaller, focused modules
  - Moved common types and interfaces to appropriate packages
  - Improved package structure
- Configuration
  - Added proper validation and defaults
  - Added environment variable support
  - Added configuration types and methods
- Logging
  - Added proper log levels and fields
  - Added proper log context and formatting
  - Added log configuration options
- Metrics and Monitoring
  - Added proper metrics collection
  - Added health checks and monitoring
  - Added metrics endpoint
- Security
  - Added proper TLS configuration
  - Added authentication and authorization
  - Added security headers

### In Progress
- Code cleanup and reorganization
  - [x] Identified duplicate test utilities
  - [x] Identified unused packages
  - [x] Identified scattered types
  - [x] Created cleanup plan
  - [ ] Started consolidation of test utilities
  - [ ] Started removal of unused code
  - [ ] Started standardization of package structure
- HTTP client error handling improvements

### Next Up
- Consolidate test utilities
- Remove unused packages
- Standardize package structure
- Clean up documentation

## Notes
- Each task should be completed in a separate commit
- Each commit should include tests
- Each commit should include documentation updates
- Each commit should be reviewed before merging
- Each commit should follow the project's coding standards
- When removing code:
  - Verify it's not used by other packages
  - Update tests to remove dependencies
  - Update documentation to reflect changes
  - Keep git history clean with atomic commits

# TODO List

## High Priority
- [x] Consolidate test utilities
  - [x] Move test utilities to `internal/testutils`
  - [x] Remove old testutils directory
  - [x] Update imports in test files

- [x] Consolidate types
  - [x] Move Job and Item structs to `internal/models`
  - [x] Move storage interface to `internal/storage`
  - [x] Remove old types packages
  - [x] Update imports in all files

- [ ] Remove unused code
  - [ ] Remove unused interfaces
  - [ ] Remove unused types
  - [ ] Remove unused functions
  - [ ] Remove unused imports

## Medium Priority
- [ ] Improve error handling
  - [ ] Add error wrapping
  - [ ] Add error context
  - [ ] Add error logging
  - [ ] Add error recovery

- [ ] Improve logging
  - [ ] Add structured logging
  - [ ] Add log levels
  - [ ] Add log rotation
  - [ ] Add log filtering

- [ ] Improve testing
  - [ ] Add more unit tests
  - [ ] Add integration tests
  - [ ] Add performance tests
  - [ ] Add load tests

## Low Priority
- [ ] Improve documentation
  - [ ] Add API documentation
  - [ ] Add usage examples
  - [ ] Add architecture diagrams
  - [ ] Add deployment guide

- [ ] Improve monitoring
  - [ ] Add metrics
  - [ ] Add tracing
  - [ ] Add alerts
  - [ ] Add dashboards

- [ ] Improve deployment
  - [ ] Add Docker support
  - [ ] Add Kubernetes support
  - [ ] Add CI/CD pipeline
  - [ ] Add release automation

### 12. Remove Named Dependencies
- [ ] Remove named fx dependencies across codebase
  - [x] Audit all fx.Module declarations in testutils
  - [x] Remove fx.Annotate usage from testutils
  - [x] Update struct field tags in testutils
  - [x] Simplify dependency providers in testutils
  - [x] Update tests to use type-based injection in testutils
  - [ ] Audit all fx.Module declarations in article module
  - [ ] Remove fx.Annotate usage from article module
  - [ ] Update struct field tags in article module
  - [ ] Simplify dependency providers in article module
  - [ ] Update tests to use type-based injection in article module
  - [ ] Verify all modules work together

Steps for each module:
1. Remove name tags from struct fields
2. Remove fx.Annotate wrappers
3. Simplify fx.Provide calls
4. Update tests to use type-based injection
5. Verify module integration

Priority order:
1. [x] cmd/common/testutils
2. [ ] internal/article
3. [ ] internal/crawler
4. [ ] internal/storage
5. [ ] internal/sources
6. [ ] internal/config
7. [ ] internal/logger
