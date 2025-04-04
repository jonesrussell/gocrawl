# GoCrawl Cleanup Plan

## High Priority

### 1. Code Cleanup and Reorganization
- [ ] Standardize Package Structure
  - [ ] Move all tests to `_test` packages
  - [ ] Add proper package documentation
  - [ ] Add proper function documentation
  - [ ] Add proper type documentation
  - [ ] Add proper interface documentation
  - [ ] Add proper examples

### 2. Config Package Improvements
- [x] Split module.go into focused files
  - [x] Create viper.go for Viper setup
  - [x] Create server.go for server config
  - [x] Create elasticsearch.go for ES config
  - [x] Create crawler.go for crawler config
  - [x] Create sources.go for sources config
  - [x] Create logging.go for logging config
  - [x] Move constants to constants.go
  - [ ] Add proper error types in errors.go
  - [ ] Add configuration versioning
  - [ ] Add schema validation
  - [ ] Add hot reload support
- [x] Split config_test.go into focused test files
  - [x] Create app_test.go for app configuration tests
  - [x] Create crawler_test.go for crawler configuration tests
  - [x] Create elasticsearch_test.go for Elasticsearch configuration tests
  - [x] Create sources_test.go for sources configuration tests
  - [x] Create logging_test.go for logging configuration tests
  - [x] Create server_test.go for server configuration tests
  - [x] Create loader_test.go for configuration loading tests
  - [x] Create module_test.go for dependency injection tests
  - [x] Create transport_test.go for HTTP transport tests
  - [x] Create priority_test.go for configuration priority tests
  - [x] Create testutils/logger_test.go for test utilities
  - [x] Move remaining test utilities to testutils package
- [ ] Enhance Testing
  - [ ] Add edge cases to TestParseRateLimit
  - [ ] Add configuration validation tests
  - [ ] Add benchmarks for performance-critical functions
  - [ ] Add concurrent access tests
  - [ ] Add hot reload tests
- [ ] Improve Error Handling
  - [ ] Add specific error types
  - [ ] Add error wrapping for Viper errors
  - [ ] Add context to error messages
  - [ ] Add validation error details
- [ ] Enhance Documentation
  - [ ] Add usage examples
  - [ ] Add configuration examples
  - [ ] Add error handling examples
  - [ ] Add hot reload documentation
- [ ] Add Monitoring
  - [ ] Add metrics for config changes
  - [ ] Add metrics for validation errors
  - [ ] Add metrics for hot reload events
  - [ ] Add tracing for config operations

### 3. HTTP Client Error Handling
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

### 4. Module Reorganization
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

### 5. Test Utilities Enhancement
- [x] Implement MockProcessor for testing
  - [x] Add Process method
  - [x] Add CanProcess method
  - [x] Add ContentType method
  - [x] Add GetMetrics method
  - [x] Add ProcessHTML method
  - [x] Add ProcessJob method
  - [x] Add Start method
  - [x] Add Stop method
  - [x] Add constructor and provider functions
- [ ] Add test utilities for other interfaces
  - [ ] Create MockStorage for storage tests
  - [ ] Create MockMetrics for metrics tests
  - [ ] Create MockLogger for logging tests
  - [ ] Create MockEventBus for event bus tests
- [ ] Add test helpers for common scenarios
  - [ ] Add helper for creating test jobs
  - [ ] Add helper for creating test articles
  - [ ] Add helper for creating test HTML content
  - [ ] Add helper for creating test configurations
- [ ] Add test fixtures
  - [ ] Add HTML fixtures for content processing
  - [ ] Add configuration fixtures
  - [ ] Add job fixtures
  - [ ] Add article fixtures

## Medium Priority

### 1. Documentation
- [ ] Add proper examples
  - [ ] Add usage examples
  - [ ] Add test examples
  - [ ] Add error handling examples
  - [ ] Add configuration examples

### 2. Testing
- [ ] Add more unit tests
  - [ ] Add tests for edge cases
  - [ ] Add tests for error conditions
  - [ ] Add tests for concurrent operations
  - [ ] Add tests for lifecycle events
- [ ] Add integration tests
  - [ ] Add tests for module interactions
  - [ ] Add tests for end-to-end workflows
  - [ ] Add tests for configuration changes
  - [ ] Add tests for error recovery
- [ ] Add performance tests
  - [ ] Add benchmarks for critical paths
  - [ ] Add load tests for concurrent operations
  - [ ] Add stress tests for error conditions
  - [ ] Add memory leak tests

### 3. Monitoring
- [ ] Add metrics
  - [ ] Add counter metrics for operations
  - [ ] Add gauge metrics for resource usage
  - [ ] Add histogram metrics for latencies
  - [ ] Add summary metrics for distributions
- [ ] Add tracing
  - [ ] Add trace points for critical paths
  - [ ] Add span attributes for context
  - [ ] Add trace sampling configuration
  - [ ] Add trace export configuration
- [ ] Add alerts
  - [ ] Add alert rules for error rates
  - [ ] Add alert rules for latency thresholds
  - [ ] Add alert rules for resource usage
  - [ ] Add alert rules for health checks
- [ ] Add dashboards
  - [ ] Add operational dashboards
  - [ ] Add performance dashboards
  - [ ] Add error dashboards
  - [ ] Add resource dashboards

## Low Priority

### 1. Deployment
- [ ] Add Docker support
  - [ ] Add Dockerfile
  - [ ] Add docker-compose.yml
  - [ ] Add Docker documentation
  - [ ] Add Docker tests
- [ ] Add Kubernetes support
  - [ ] Add Kubernetes manifests
  - [ ] Add Helm charts
  - [ ] Add Kubernetes documentation
  - [ ] Add Kubernetes tests
- [ ] Add CI/CD pipeline
  - [ ] Add GitHub Actions workflow
  - [ ] Add automated testing
  - [ ] Add automated deployment
  - [ ] Add release automation
- [ ] Add release automation
  - [ ] Add version management
  - [ ] Add changelog generation
  - [ ] Add release notes
  - [ ] Add release verification

## Best Practices

### 1. Code Organization
- Keep related functionality together
- Use consistent naming across related components
- Avoid type stuttering in package names
- Use descriptive variable names
- Avoid global state
- Keep function complexity below 30
- Keep package average complexity below 10
- Limit function length to 100 lines or 50 statements
- Use canonical import paths

### 2. Dependency Injection
- Define interfaces in the package that uses them
- For module-specific interfaces, define them in module.go
- For interfaces used across multiple packages, define them in the consuming package
- Avoid interface stuttering
- Define New* constructors only in module.go files
- When testing modules:
  - Provide all required dependencies
  - Avoid providing dependencies that are already provided
  - Use test helpers for common dependency setup
  - Verify dependency injection before testing functionality

### 3. Error Handling
- Prefer errors.New over fmt.Errorf for simple error messages
- Check error return values
- Use descriptive error variable names
- Wrap errors with context when appropriate
- Always check type assertion errors
- Check SQL Row and Statement closures
- Include context in error handling where available

### 4. Testing
- Use require for error assertions
- Avoid allocations with (*os.File).WriteString
- Avoid shadow declarations
- Test each module independently
- Write testable examples with expected output
- Use t.Parallel() appropriately in tests
- Place tests in a separate _test package
- Use testify consistently for assertions
- Follow test naming conventions:
  - Test[FunctionName] for function tests
  - Test[TypeName] for type tests
  - Test[InterfaceName] for interface tests
- Structure tests with setup, execution, and verification phases
- Use table-driven tests for multiple test cases
- Mock external dependencies consistently
- Use test helpers to reduce code duplication
- Test both success and failure cases
- Test edge cases and boundary conditions
- Use meaningful test data
- Document test setup and assumptions
- Keep tests focused and single-purpose
- Use subtests for related test cases
- Test public interfaces, not implementation details
- Use test fixtures for complex test data
- Clean up resources in test teardown
- Use test-specific types and interfaces
- Avoid test interdependence
- Use meaningful assertion messages
- Test error conditions explicitly
- Use test coverage tools appropriately

### 5. Documentation
- Provide brief explanations of code updates
- Document interface requirements
- Include examples for complex patterns
- End comments with periods
- Document magic numbers with constants
- Include context in logging calls where available

### 6. Style
- Use named returns sparingly and document when used
- Ensure exhaustive switch statements and map initializations
- Include context in HTTP requests
- Close response bodies in HTTP clients
- Use standard library variables where available
- Avoid magic numbers in code

### 7. HTML Processing
- Use goquery for HTML parsing
- Validate HTML before processing
- Handle malformed HTML gracefully
- Use CSS selectors for element selection
- Document selector requirements
- Handle timezone conversions explicitly
- Validate time formats before parsing
- Use constants for common time formats
- Handle empty or missing elements gracefully
- Document error conditions and recovery
- Use meaningful error messages
- Include context in error handling
- Clean up resources properly
- Handle concurrent access safely
- Use appropriate logging levels
- Document performance considerations
- Use appropriate metrics for monitoring
- Handle edge cases explicitly
- Use table-driven tests for validation
- Mock external dependencies in tests
- Use test fixtures for complex HTML
- Document test assumptions
- Clean up test resources
- Use appropriate test helpers
- Test error conditions thoroughly
- Use meaningful test data
- Document test setup
- Keep tests focused
- Use subtests for related cases
- Test public interfaces
- Use test-specific types
- Avoid test interdependence
- Use meaningful assertions
- Test error conditions
- Use coverage tools

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