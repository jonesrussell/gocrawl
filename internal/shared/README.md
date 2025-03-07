# Shared Package

The shared package provides global variables and functions that need to be accessed across different parts of the GoCrawl application. While global state should generally be avoided, this package provides a controlled way to share essential application-wide resources.

## Components

### Globals (`globals.go`)
Manages global application resources and their initialization.

Key features:
- Global configuration access
- Global logger access
- Controlled initialization
- Thread-safe access
- Centralized management
- Consistent usage

## Usage Examples

### Setting Global Resources
```go
// Set global configuration
cfg := &config.Config{...}
shared.SetConfig(cfg)

// Set global logger
log := logger.New(...)
shared.SetLogger(log)
```

### Accessing Global Resources
```go
// Access configuration
baseURL := shared.Config.Crawler.BaseURL
maxDepth := shared.Config.Crawler.MaxDepth

// Use logger
shared.Logger.Info("Starting operation")
shared.Logger.Debug("Processing item", "id", itemID)
shared.Logger.Error("Operation failed", "error", err)
```

## Key Features

### 1. Global Resource Management
- Centralized configuration access
- Structured logging capabilities
- Controlled initialization
- Thread-safe operations
- Consistent access patterns
- Resource lifecycle management

### 2. Configuration Access
- Application settings
- Environment variables
- Feature flags
- Connection settings
- Runtime parameters
- Default values

### 3. Logging Capabilities
- Structured logging
- Error reporting
- Debug information
- Operation tracking
- Performance monitoring
- Audit logging

### 4. Resource Safety
- Controlled initialization
- Thread-safe access
- Resource validation
- Error handling
- State management
- Lifecycle control

## Best Practices

1. **Initialization**
   - Initialize early in startup
   - Validate all resources
   - Handle initialization errors
   - Set up in correct order
   - Document dependencies
   - Verify initialization

2. **Resource Access**
   - Use consistent patterns
   - Check for nil values
   - Handle missing resources
   - Log access errors
   - Follow thread safety
   - Document usage

3. **Configuration**
   - Set before access
   - Validate settings
   - Handle missing values
   - Use type safety
   - Document requirements
   - Provide defaults

4. **Logging**
   - Use structured logging
   - Include context
   - Handle errors
   - Follow log levels
   - Add timestamps
   - Include metadata

## Development

When modifying the shared package:
1. Minimize global state
2. Document thread safety
3. Handle initialization
4. Validate resources
5. Test thoroughly
6. Consider dependencies
7. Update documentation
8. Handle errors properly

## Important Notes

1. **Global State**
   - Use sparingly
   - Document necessity
   - Consider alternatives
   - Handle thread safety
   - Manage lifecycle
   - Control access

2. **Thread Safety**
   - Initialize safely
   - Access atomically
   - Handle concurrency
   - Protect resources
   - Document patterns
   - Test thoroughly

3. **Error Handling**
   - Check for nil
   - Validate state
   - Log errors
   - Provide context
   - Handle gracefully
   - Document recovery

4. **Documentation**
   - Explain purpose
   - Document patterns
   - Show examples
   - Note limitations
   - Update changes
   - Include warnings 