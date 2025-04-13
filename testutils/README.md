# Test Utilities

This directory contains test utilities and mocks for the GoCrawl project.

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