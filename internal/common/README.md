# Common Package

The common package provides shared functionality, constants, and utilities used across the GoCrawl application. It serves as a central location for common types, interfaces, and helper functions.

## Components

### Constants (`constants.go`)
Defines default timeout values used throughout the application.

Key features:
- Default shutdown timeout
- Default startup timeout
- Default operation timeout
- Configurable through overrides

### Module (`module.go`)
Manages dependency injection and shared dependencies using the fx framework.

Key features:
- Type aliases for core interfaces
- Centralized module configuration
- Dependency injection setup
- Core service integration
- Logging configuration

### Output (`output.go`)
Provides consistent output formatting and user interaction utilities.

Key features:
- Error message formatting
- Success message formatting
- Information message display
- User confirmation prompts
- Table formatting utilities
- Visual dividers

## Type Aliases

The package provides type aliases for commonly used interfaces and types:

```go
// Logger is an alias for the logger interface
var logger common.Logger

// Storage is an alias for the storage interface
var storage common.Storage

// Config is an alias for the configuration type
var config common.Config
```

## Module

The common module provides shared dependencies for commands:

```go
var Module = fx.Module("common",
    config.Module,
    logger.Module,
    storage.Module,
)
```

## Usage

Import the common package to access shared types and utilities:

```go
import "github.com/jonesrussell/gocrawl/internal/common"

func example(logger common.Logger, storage common.Storage) {
    // Use logger and storage
}
```

## Usage Examples

### Output Formatting
```go
// Display error message
common.PrintErrorf("Failed to connect: %v", err)

// Display success message
common.PrintSuccessf("Successfully processed %d items", count)

// Display information
common.PrintInfof("Processing items...")

// Get user confirmation
if common.PrintConfirmation("Do you want to continue?") {
    // User confirmed
}

// Print table header
common.PrintTableHeaderf("%-20s %-10s %s", "Name", "Status", "Description")

// Print divider
common.PrintDivider(50)
```

### Using Timeouts
```go
// Use default timeouts
ctx, cancel := context.WithTimeout(context.Background(), common.DefaultOperationTimeout)
defer cancel()

// Use shutdown timeout
srv.Shutdown(context.WithTimeout(context.Background(), common.DefaultShutdownTimeout))
```

## Key Features

### 1. Output Formatting
- Consistent error messages
- Visual success indicators
- Information display
- User interaction
- Table formatting
- Visual separators

### 2. Dependency Management
- Interface aliases
- Type definitions
- Module configuration
- Service integration
- Logging setup

### 3. Constants
- Timeout definitions
- Default values
- Configuration options
- Common settings
- Overridable defaults

### 4. Error Handling
- Error formatting
- User feedback
- Status updates
- Operation results
- Visual indicators

## Best Practices

1. **Output Formatting**
   - Use appropriate message types
   - Maintain consistent formatting
   - Include relevant details
   - Handle errors gracefully
   - Provide clear feedback
   - Use visual indicators

2. **Dependency Management**
   - Use type aliases
   - Follow DI patterns
   - Maintain modularity
   - Handle dependencies properly
   - Document interfaces
   - Keep dependencies minimal

3. **Timeouts**
   - Use appropriate timeouts
   - Handle context cancellation
   - Implement graceful shutdown
   - Monitor long operations
   - Provide override options
   - Document timeout usage

4. **Error Handling**
   - Use consistent formatting
   - Include context
   - Provide clear messages
   - Handle all cases
   - Log appropriately
   - Give user feedback

## Development

When modifying the common package:
1. Update documentation
2. Maintain backward compatibility
3. Follow error handling patterns
4. Keep formatting consistent
5. Update type aliases
6. Test thoroughly
7. Consider dependencies
8. Handle edge cases 