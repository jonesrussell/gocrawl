# Sources Package TODO

## Current Issues
1. Mixed responsibilities:
   - Source configuration management
   - Crawler instance management
   - Index management
   - Logger dependency

2. Direct dependencies that should be abstracted:
   - Crawler interface
   - Index manager
   - Logger interface

## Core Responsibilities
The sources package should focus solely on:
1. Source Configuration Management
   - Loading source configurations
   - Validating configurations
   - Providing access to source definitions
2. Source Lifecycle Management
   - Source state tracking
   - Configuration updates
   - Source validation

## Proposed Changes

### 1. Interface Definition
```go
type SourceManager interface {
    // GetSource returns a source configuration by name
    GetSource(name string) (*Config, error)
    
    // ListSources returns all available sources
    ListSources() []*Config
    
    // ValidateSource validates a source configuration
    ValidateSource(config *Config) error
    
    // ReloadSources reloads source configurations from disk
    ReloadSources() error
}
```

### 2. Configuration Structure
```go
type Config struct {
    Name        string            `yaml:"name"`
    URL         string            `yaml:"url"`
    RateLimit   string            `yaml:"rate_limit"`
    MaxDepth    int              `yaml:"max_depth"`
    Selectors   SelectorConfig    `yaml:"selectors"`
    Metadata    map[string]string `yaml:"metadata,omitempty"`
}
```

### 3. New Package Structure
```
sources/
├── module.go       # DI setup
├── sources.go      # Core implementation
├── config.go       # Configuration types
├── validator.go    # Configuration validation
└── errors.go       # Package-specific errors
```

## Implementation Tasks
1. [ ] Remove crawler dependency
2. [ ] Remove index manager dependency
3. [ ] Move validation logic to separate file
4. [ ] Implement configuration reloading
5. [ ] Add configuration validation
6. [ ] Update tests for new structure
7. [ ] Add proper error types
8. [ ] Update documentation

## Integration with CLI Commands
1. [ ] Update fx providers in module.go
2. [ ] Ensure proper lifecycle hooks for config loading
3. [ ] Add CLI-specific validation methods
4. [ ] Implement source listing for CLI commands
5. [ ] Add source validation command
6. [ ] Support configuration hot-reloading

## Migration Plan
1. Create new interfaces in common package
2. Remove crawler and index dependencies
3. Update module providers
4. Update CLI command integration
5. Update tests
6. Update documentation

## Dependencies
- gopkg.in/yaml.v3 for configuration
- go.uber.org/fx for DI
- go.uber.org/zap for logging

## Notes
- Keep configuration loading synchronous
- Support both file and environment-based configuration
- Add proper validation error messages
- Support source-specific metadata
- Add configuration schema validation
- Support configuration inheritance
- Add configuration versioning

## High Priority

### Error Handling
- [ ] Add more specific error types for configuration issues
- [ ] Implement validation for all YAML fields
- [ ] Add error recovery for failed source loading
- [ ] Improve error context in logging
- [ ] Add validation for rate limit format
- [ ] Add validation for URL format
- [ ] Add validation for selector patterns

### Performance
- [ ] Implement source configuration caching
- [ ] Add parallel source processing
- [ ] Optimize source lookup
- [ ] Add metrics collection
- [ ] Implement source health checks
- [ ] Add performance monitoring
- [ ] Optimize YAML parsing

### Configuration
- [ ] Add support for environment variables
- [ ] Implement configuration hot-reloading
- [ ] Add support for multiple config files
- [ ] Add support for config templates
- [ ] Add validation for required fields
- [ ] Add support for config inheritance
- [ ] Add support for config overrides

## Medium Priority

### Testing
- [ ] Add more integration tests
- [ ] Add performance benchmarks
- [ ] Add fuzzing tests for YAML parsing
- [ ] Add mock implementations
- [ ] Add tests for error cases
- [ ] Add tests for validation
- [ ] Add tests for hot-reloading
- [ ] Add tests for inheritance

### Documentation
- [ ] Add more configuration examples
- [ ] Document validation rules
- [ ] Add troubleshooting guide
- [ ] Document error codes
- [ ] Add architecture diagrams
- [ ] Document debugging procedures
- [ ] Add configuration best practices
- [ ] Document selector patterns

### Features
- [ ] Add support for source groups
- [ ] Implement source dependencies
- [ ] Add support for source scheduling
- [ ] Add support for source priorities
- [ ] Add support for source tags
- [ ] Implement source filtering
- [ ] Add support for source templates
- [ ] Add support for source inheritance

## Low Priority

### Code Organization
- [ ] Split large files into smaller components
- [ ] Add more interface abstractions
- [ ] Improve dependency injection
- [ ] Add more logging points
- [ ] Refactor configuration handling
- [ ] Improve error handling structure
- [ ] Add more modular components
- [ ] Improve code reusability

### Monitoring
- [ ] Add health check endpoints
- [ ] Implement detailed metrics
- [ ] Add performance profiling
- [ ] Add resource usage tracking
- [ ] Add source status tracking
- [ ] Add configuration metrics
- [ ] Add error rate tracking
- [ ] Add success rate tracking

### Development Experience
- [ ] Add development tools
- [ ] Improve error messages
- [ ] Add more debug options
- [ ] Add development mode
- [ ] Add debugging helpers
- [ ] Improve logging format
- [ ] Add development utilities
- [ ] Add testing helpers

## Future Considerations

### Scalability
- [ ] Add distributed source management
- [ ] Implement source load balancing
- [ ] Add support for source clusters
- [ ] Add support for source federation
- [ ] Add horizontal scaling
- [ ] Implement work distribution
- [ ] Add resource sharing
- [ ] Add state management

### Integration
- [ ] Add support for more config formats
- [ ] Add support for remote configs
- [ ] Add support for config validation
- [ ] Add support for config migration
- [ ] Add API integration
- [ ] Add webhook support
- [ ] Add event streaming
- [ ] Add data export

### Security
- [ ] Add support for encrypted configs
- [ ] Implement config signing
- [ ] Add support for access control
- [ ] Add support for audit logging
- [ ] Add request signing
- [ ] Add content validation
- [ ] Add input sanitization
- [ ] Add output sanitization 