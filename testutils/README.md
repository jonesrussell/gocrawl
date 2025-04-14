# Test Utilities

This directory contains shared test utilities and mock implementations used across the codebase.

## Directory Structure

```
/testutils/
├── mocks/           # Generated mock implementations
│   ├── api/        # API-related mocks
│   ├── config/     # Config-related mocks
│   ├── crawler/    # Crawler-related mocks
│   ├── logger/     # Logger-related mocks
│   ├── sources/    # Sources-related mocks
│   ├── models/     # Models-related mocks
│   └── storage/    # Storage-related mocks
├── errors.go       # Common error test utilities
├── security.go     # Security test utilities
└── storage.go      # Storage test utilities
```

## Usage

Import test utilities using the full package path:

```go
import (
    "github.com/jonesrussell/gocrawl/testutils"
    "github.com/jonesrussell/gocrawl/testutils/mocks/api"
)
```

## Package-Specific Test Helpers

Package-specific test helpers should be placed in their respective packages with a `test_` prefix:

```go
// internal/config/test_config.go
package config

// TestConfig creates a test configuration
func TestConfig() *Config {
    // ...
}
```

## Mock Generation

Mocks are generated using the `scripts/generate_mocks.sh` script. This script:
1. Creates the necessary mock directories
2. Generates mocks for all interfaces
3. Formats the generated code

To generate mocks:
```bash
./scripts/generate_mocks.sh
```

## Mock Structure

Mocks are organized by package to make them easy to find and maintain:

```
/testutils/
  ├── mocks/
  │   ├── api/           # API-related mocks
  │   ├── config/        # Configuration mocks
  │   ├── crawler/       # Crawler mocks
  │   ├── indices/       # Index-related mocks
  │   ├── models/        # Model mocks
  │   └── storage/       # Storage mocks
  └── README.md
```

## Usage

To use a mock in your tests:

1. Import the mock package:
   ```go
   import "github.com/jonesrussell/gocrawl/testutils/mocks/api"
   ```

2. Create a mock instance:
   ```go
   mockIndexManager := &mocks.IndexManager{}
   ```

3. Set up expectations:
   ```go
   mockIndexManager.On("CreateIndex", mock.Anything).Return(nil)
   ```

## Best Practices

1. Keep mocks focused and minimal
2. Use descriptive names for mock methods
3. Document mock behavior in comments
4. Update mocks when interfaces change
5. Use testify/mock for consistency 