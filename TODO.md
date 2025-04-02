# GoCrawl Cleanup Plan

## High Priority

### 1. Dependency Injection Simplification
- [ ] Simplify Logger Implementation
  - [ ] Consolidate logger interfaces into single package
    - [ ] Move logger interface to internal/logger
    - [ ] Remove common/types logger interface
    - [ ] Update all imports to use new logger package
  - [ ] Simplify zap integration
    - [ ] Remove unnecessary wrapper layers
    - [ ] Use zap directly with minimal abstraction
    - [ ] Keep only essential logging methods
  - [ ] Remove fx module for logger
    - [ ] Create simple constructor function
    - [ ] Use environment-based configuration
    - [ ] Add proper error handling
  - [ ] Update tests
    - [ ] Simplify mock logger implementation
    - [ ] Update test utilities
    - [ ] Add proper test coverage
- [ ] Remove Unnecessary Abstractions
  - [ ] Remove common/types package
    - [ ] Move interfaces to consuming packages
    - [ ] Update all imports
    - [ ] Update tests
  - [ ] Remove type aliases
    - [ ] Use direct type references
    - [ ] Update imports
    - [ ] Update tests
  - [ ] Simplify module structure
    - [ ] Reduce nesting levels
    - [ ] Combine related modules
    - [ ] Remove unnecessary modules
- [ ] Simplify Dependency Injection
  - [ ] Remove named dependencies
    - [ ] Use type-based injection
    - [ ] Update fx modules
    - [ ] Update tests
  - [ ] Use constructor injection
    - [ ] Replace fx where appropriate
    - [ ] Add proper error handling
    - [ ] Update tests
  - [ ] Keep fx for application composition only
    - [ ] Remove unnecessary fx usage
    - [ ] Simplify module structure
    - [ ] Update tests
- [ ] Review and Replace Cobra CLI
  - [ ] Evaluate current CLI needs
    - [ ] List required commands
    - [ ] Document usage patterns
    - [ ] Identify complexity points
  - [ ] Design simpler CLI
    - [ ] Use standard flag package
    - [ ] Create simple command structure
    - [ ] Add proper help messages
  - [ ] Implement new CLI
    - [ ] Create basic command structure
    - [ ] Add command implementations
    - [ ] Add proper error handling
  - [ ] Update tests
    - [ ] Add CLI tests
    - [ ] Add integration tests
    - [ ] Update existing tests
  - [ ] Remove Cobra dependency
    - [ ] Clean up imports
    - [ ] Update documentation
    - [ ] Update build scripts

### 2. Module Reorganization
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
  - [ ] Create separate sources module
    - [x] Move sources interface to internal/sources
      - [x] Create internal/sources/interface.go
        - [x] Define Interface type with required methods
        - [x] Add interface documentation
        - [x] Add interface examples
        - [x] Add interface validation
      - [x] Create internal/sources/types.go
        - [x] Define Config struct
        - [x] Define Params struct
        - [x] Add type documentation
        - [x] Add type validation
      - [x] Create internal/sources/errors.go
        - [x] Define error types
        - [x] Add error documentation
        - [x] Add error examples
    - [ ] Create sources module for dependency injection
      - [ ] Create internal/sources/module.go
        - [ ] Define Module variable
        - [ ] Add module documentation
        - [ ] Add module examples
        - [ ] Add module validation
      - [ ] Create internal/sources/loader.go
        - [ ] Implement source loading
        - [ ] Add loader documentation
        - [ ] Add loader examples
        - [ ] Add loader validation
    - [ ] Add sources tests
      - [x] Create internal/sources/sources_test.go
        - [x] Add interface tests
        - [x] Add module tests
        - [x] Add integration tests
      - [ ] Create internal/sources/loader_test.go
        - [ ] Add loader tests
        - [ ] Add validation tests
        - [ ] Add error tests
    - [ ] Update commands to use new sources module
      - [ ] Update job command
        - [ ] Update dependencies
        - [ ] Update tests
        - [ ] Update documentation
      - [ ] Update crawl command
        - [ ] Update dependencies
        - [ ] Update tests
        - [ ] Update documentation
      - [ ] Update httpd command
        - [ ] Update dependencies
        - [ ] Update tests
        - [ ] Update documentation
    - [ ] Combine internal/sources and internal/sources
      - [x] Move unique functionality from internal/sources to internal/sources
        - [x] Move interface definitions
        - [x] Move type definitions
        - [x] Move error definitions
        - [x] Move loader implementation
      - [x] Update all references to use internal/sources
        - [x] Update imports in all files
        - [x] Update type references
        - [x] Update interface references
      - [x] Delete internal/sources package
        - [x] Remove all files
        - [x] Update documentation
      - [x] Update test files
        - [x] Update imports
        - [x] Update type references
        - [x] Update interface references
        - [x] Update mock implementations
      - [ ] Verify all functionality works
        - [ ] Run all tests
        - [ ] Check coverage
        - [ ] Verify no regressions
    - [ ] Add HTTP client error handling
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
    - [ ] Add examples
      - [ ] Add usage examples
        - [ ] Add basic usage
        - [ ] Add advanced usage
        - [ ] Add error handling
      - [ ] Add test examples
        - [ ] Add unit tests
        - [ ] Add integration tests
        - [ ] Add mock examples
      - [ ] Add error handling examples
        - [ ] Add error types
        - [ ] Add error wrapping
        - [ ] Add error recovery
      - [ ] Add configuration examples
        - [ ] Add basic config
        - [ ] Add advanced config
        - [ ] Add validation
  - [ ] Create separate storage module
    - [ ] Move storage interface to pkg/storage
      - [ ] Create pkg/storage/interface.go
      - [ ] Define Interface type
      - [ ] Add interface documentation
      - [ ] Add interface examples
    - [ ] Create storage module for dependency injection
      - [ ] Create pkg/storage/module.go
      - [ ] Define Module variable
      - [ ] Add module documentation
      - [ ] Add module examples
    - [ ] Add storage tests
      - [ ] Create pkg/storage/storage_test.go
      - [ ] Add interface tests
      - [ ] Add module tests
      - [ ] Add integration tests
    - [ ] Update commands to use new storage module
      - [ ] Update job command
      - [ ] Update crawl command
      - [ ] Update httpd command
  - [ ] Create separate metrics module
    - [ ] Move metrics interface to pkg/metrics
      - [ ] Create pkg/metrics/interface.go
      - [ ] Define Interface type
      - [ ] Add interface documentation
      - [ ] Add interface examples
    - [ ] Create metrics module for dependency injection
      - [ ] Create pkg/metrics/module.go
      - [ ] Define Module variable
      - [ ] Add module documentation
      - [ ] Add module examples
    - [ ] Add metrics tests
      - [ ] Create pkg/metrics/metrics_test.go
      - [ ] Add interface tests
      - [ ] Add module tests
      - [ ] Add integration tests
    - [ ] Update commands to use new metrics module
      - [ ] Update job command
      - [ ] Update crawl command
      - [ ] Update httpd command
- [ ] Make dependencies explicit in each module
  - [x] Remove implicit dependencies from common module
    - [x] Remove logger dependency from common module
    - [x] Remove config dependency from common module
    - [ ] Remove sources dependency from common module
      - [ ] Remove sources imports
      - [ ] Remove sources types
      - [ ] Remove sources functions
      - [ ] Update tests
    - [ ] Remove storage dependency from common module
      - [ ] Remove storage imports
      - [ ] Remove storage types
      - [ ] Remove storage functions
      - [ ] Update tests
    - [ ] Remove metrics dependency from common module
      - [ ] Remove metrics imports
      - [ ] Remove metrics types
      - [ ] Remove metrics functions
      - [ ] Update tests
  - [x] Add explicit dependency declarations in each module
    - [x] Add logger dependency declarations
    - [x] Add config dependency declarations
    - [ ] Add sources dependency declarations
      - [ ] Add logger dependency
      - [ ] Add config dependency
      - [ ] Add validation
      - [ ] Add tests
    - [ ] Add storage dependency declarations
      - [ ] Add logger dependency
      - [ ] Add config dependency
      - [ ] Add validation
      - [ ] Add tests
    - [ ] Add metrics dependency declarations
      - [ ] Add logger dependency
      - [ ] Add config dependency
      - [ ] Add validation
      - [ ] Add tests
  - [x] Update module tests to reflect explicit dependencies
    - [x] Update logger module tests
    - [x] Update config module tests
    - [ ] Update sources module tests
    - [ ] Update storage module tests
    - [ ] Update metrics module tests
  - [ ] Add dependency validation in each module
    - [x] Add logger dependency validation
    - [x] Add config dependency validation
    - [ ] Add sources dependency validation
    - [ ] Add storage dependency validation
    - [ ] Add metrics dependency validation
- [ ] Implement interface segregation
  - [x] Create specific interfaces for each module
    - [x] Create logger interface
    - [x] Create config interface
    - [ ] Create sources interface
    - [ ] Create storage interface
    - [ ] Create metrics interface
  - [x] Move interfaces to consuming packages
    - [x] Move logger interface to internal/logger
    - [x] Move config interface to internal/config
    - [ ] Move sources interface to internal/sources
    - [ ] Move storage interface to pkg/storage
    - [ ] Move metrics interface to pkg/metrics
  - [ ] Remove generic interfaces from common package
    - [x] Remove logger interface from common package
    - [x] Remove config interface from common package
    - [ ] Remove sources interface from common package
    - [ ] Remove storage interface from common package
    - [ ] Remove metrics interface from common package
  - [ ] Update tests to use specific interfaces
    - [x] Update tests to use logger interface
    - [x] Update tests to use config interface
    - [ ] Update tests to use sources interface
    - [ ] Update tests to use storage interface
    - [ ] Update tests to use metrics interface
- [ ] Reorganize shared code
  - [x] Move shared code from internal/common to pkg
    - [x] Move logger code to internal/logger
    - [x] Move config code to internal/config
    - [ ] Move sources code to internal/sources
    - [ ] Move storage code to pkg/storage
    - [ ] Move metrics code to pkg/metrics
  - [x] Create internal/logger for logging utilities
    - [x] Create logger interface
    - [x] Create logger module
    - [x] Create logger tests
  - [x] Create internal/config for configuration utilities
    - [x] Create config interface
    - [x] Create config module
    - [x] Create config tests
  - [ ] Create pkg/metrics for metrics utilities
    - [ ] Create metrics interface
    - [ ] Create metrics module
    - [ ] Create metrics tests
  - [ ] Create pkg/storage for storage utilities
    - [ ] Create storage interface
    - [ ] Create storage module
    - [ ] Create storage tests
  - [ ] Update imports across the codebase
    - [x] Update logger imports
    - [x] Update config imports
    - [ ] Update sources imports
    - [ ] Update storage imports
    - [ ] Update metrics imports

### 3. Interface Organization and Naming
- [x] Move interfaces to consuming packages
  - [x] Move `ContentProcessor` from `models` to `collector`
  - [x] Move `Logger` interface from `collector` to `common`
  - [x] Move `ArticleProcessor` to `article` package
- [x] Rename interfaces to avoid stuttering
  - [x] `models.ContentProcessor` → `collector.Processor`
  - [x] `collector.ArticleProcessor` → `article.Processor`
  - [x] `collector.ContentProcessor` → `content.Processor`

### 4. Dependency Injection
- [x] Improve fx module organization
  - [x] Move all fx-related code to `module.go` files
  - [x] Use `fx.Annotate` for named dependencies
  - [x] Add proper lifecycle hooks
  - [x] Add proper error handling in constructors
- [x] Add proper dependency validation
  - [x] Add validation for required dependencies
  - [x] Add validation for optional dependencies
  - [x] Add proper error messages
- [x] Fix dependency conflicts
  - [x] Remove duplicate providers
  - [x] Use mock configurations in tests
  - [x] Properly scope test modules

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
- HTTP client error handling improvements
- Example code additions
- Complete HTML processor implementation
  - [x] Add extractList method for categories and tags
  - [x] Add extractMetadata method for additional metadata
  - [x] Complete extractTime method with time format parsing
  - [x] Add tests for HTML parsing
  - [x] Add tests for time parsing
  - [x] Add tests for metadata extraction
  - [x] Add tests for error cases
  - [ ] Add metrics collection
    - [ ] Track processing time
    - [ ] Track number of elements processed
    - [ ] Track number of errors
  - [ ] Add context support
    - [ ] Allow cancellation of long-running processing
    - [ ] Support timeouts
    - [ ] Support request-scoped values
  - [ ] Add validation for configuration
    - [ ] Validate selectors
    - [ ] Validate time formats
    - [ ] Validate required fields
  - [ ] Add support for custom time formats
    - [ ] Allow configuration of additional time formats
    - [ ] Support timezone handling

### Next Up
- Add usage examples for each major component
- Add test examples demonstrating common patterns
- Add error handling examples showing best practices
- Add configuration examples for different scenarios

## Notes
- Each task should be completed in a separate commit
- Each commit should include tests
- Each commit should include documentation updates
- Each commit should be reviewed before merging
- Each commit should follow the project's coding standards

# TODO List

## High Priority
- [x] Fix dependency injection issues in API module
- [x] Fix duplicate SearchManager provider
- [x] Fix logger dependency in API module
- [x] Fix config dependency in API module
- [ ] Add command integration tests
  - [x] Test HTTP server command
    - [x] Test server startup
    - [x] Test graceful shutdown
    - [x] Test health check endpoint
    - [x] Test search endpoint
    - [x] Test error handling
  - [ ] Test crawler command
    - [ ] Test crawler startup
    - [ ] Test crawler shutdown
    - [ ] Test source validation
    - [ ] Test error handling
  - [ ] Test index command
    - [ ] Test index creation
    - [ ] Test index deletion
    - [ ] Test index mapping
    - [ ] Test error handling
  - [ ] Test dry-run command
    - [ ] Test configuration validation
    - [ ] Test source validation
    - [ ] Test error handling
  - [ ] Test version command
    - [ ] Test version output
    - [ ] Test build info output

## Medium Priority
- [ ] Add more test cases for HTML processor
  - [ ] Test malformed HTML handling
  - [ ] Test concurrent processing
  - [ ] Test memory usage
  - [ ] Test error recovery
- [ ] Improve error handling in storage module
  - [ ] Add retry mechanism for failed operations
  - [ ] Add circuit breaker for failing operations
  - [ ] Add metrics for error rates
  - [ ] Add error reporting
- [ ] Add metrics collection
  - [ ] Add Prometheus metrics
  - [ ] Add Grafana dashboards
  - [ ] Add alerting rules
  - [ ] Add monitoring documentation

## Low Priority
- [ ] Add more documentation
  - [ ] Add API documentation
  - [ ] Add configuration documentation
  - [ ] Add deployment guide
  - [ ] Add troubleshooting guide
- [ ] Add more examples
  - [ ] Add basic usage examples
  - [ ] Add advanced usage examples
  - [ ] Add integration examples
  - [ ] Add deployment examples
- [ ] Add more features
  - [ ] Add rate limiting
  - [ ] Add caching
  - [ ] Add authentication
  - [ ] Add authorization

### 13. Remove Named Dependencies
- [ ] Remove named fx dependencies across codebase
  - [ ] Audit all fx.Module declarations
  - [ ] Remove fx.Annotate usage
  - [ ] Update struct field tags
  - [ ] Simplify dependency providers
  - [ ] Update tests to use type-based injection
  - [ ] Verify all modules work together

Steps for each module:
1. Remove name tags from struct fields
2. Remove fx.Annotate wrappers
3. Simplify fx.Provide calls
4. Update tests to use type-based injection
5. Verify module integration

Priority order:
1. cmd/common/testutils
2. internal/api
3. internal/collector
4. internal/crawler
5. internal/storage
6. internal/sources
7. internal/config
8. internal/logger
