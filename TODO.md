# GoCrawl Cleanup Plan

## High Priority

### 1. Module Reorganization
- [ ] Split common module into specific modules
  - [x] Create separate logger module
    - [x] Move logger interface to pkg/logger
    - [x] Create logger module for dependency injection
    - [x] Add logger tests
    - [x] Update job command to use new logger module
  - [x] Create separate config module
    - [x] Move config interface to pkg/config
    - [x] Create config module for dependency injection
    - [x] Add config tests
    - [ ] Update commands to use new config module
      - [x] Fix type mismatches between old and new config interfaces
      - [x] Update job command to use new config module
      - [ ] Update crawl command to use new config module
      - [ ] Update httpd command to use new config module
  - [ ] Create separate sources module
    - [ ] Move sources interface to pkg/sources
    - [ ] Create sources module for dependency injection
    - [ ] Add sources tests
    - [ ] Update commands to use new sources module
  - [ ] Create separate storage module
    - [ ] Move storage interface to pkg/storage
    - [ ] Create storage module for dependency injection
    - [ ] Add storage tests
    - [ ] Update commands to use new storage module
  - [ ] Create separate metrics module
    - [ ] Move metrics interface to pkg/metrics
    - [ ] Create metrics module for dependency injection
    - [ ] Add metrics tests
    - [ ] Update commands to use new metrics module
- [ ] Make dependencies explicit in each module
  - [x] Remove implicit dependencies from common module
    - [x] Remove logger dependency from common module
    - [x] Remove config dependency from common module
    - [ ] Remove sources dependency from common module
    - [ ] Remove storage dependency from common module
    - [ ] Remove metrics dependency from common module
  - [x] Add explicit dependency declarations in each module
    - [x] Add logger dependency declarations
    - [x] Add config dependency declarations
    - [ ] Add sources dependency declarations
    - [ ] Add storage dependency declarations
    - [ ] Add metrics dependency declarations
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
    - [x] Move logger interface to pkg/logger
    - [x] Move config interface to pkg/config
    - [ ] Move sources interface to pkg/sources
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
    - [x] Move logger code to pkg/logger
    - [x] Move config code to pkg/config
    - [ ] Move sources code to pkg/sources
    - [ ] Move storage code to pkg/storage
    - [ ] Move metrics code to pkg/metrics
  - [x] Create pkg/logger for logging utilities
    - [x] Create logger interface
    - [x] Create logger module
    - [x] Create logger tests
  - [x] Create pkg/config for configuration utilities
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

### 2. Interface Organization and Naming
- [x] Move interfaces to consuming packages
  - [x] Move `ContentProcessor` from `models` to `collector`
  - [x] Move `Logger` interface from `collector` to `common`
  - [x] Move `ArticleProcessor` to `article` package
- [x] Rename interfaces to avoid stuttering
  - [x] `models.ContentProcessor` → `collector.Processor`
  - [x] `collector.ArticleProcessor` → `article.Processor`
  - [x] `collector.ContentProcessor` → `content.Processor`

### 3. Dependency Injection
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

### 4. Error Handling
- [x] Standardize error handling
  - [x] Use `errors.New` for simple errors
  - [x] Use `fmt.Errorf` with `%w` for wrapped errors
  - [x] Add proper error context
  - [x] Add proper error types
- [ ] Improve HTTP client error handling
  - [ ] Add proper response body closure
  - [ ] Add proper context handling
  - [ ] Add proper timeout handling

## Medium Priority

### 5. Testing
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

### 6. Code Organization
- [x] Split large files
  - [x] Split `collector.go`
  - [x] Split `crawler.go`
  - [x] Split `content.go`
- [x] Improve package organization
  - [x] Move common types to `common`
  - [x] Move common interfaces to `common`
  - [x] Move common constants to `common`
  - [x] Move common utilities to `common`

### 7. Documentation
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

### 8. Configuration
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

### 9. Logging
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

### 10. Metrics and Monitoring
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

### 11. Security
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
