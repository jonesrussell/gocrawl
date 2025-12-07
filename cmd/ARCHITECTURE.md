# Command Architecture

## Overview

The `cmd/` directory contains the CLI implementation for gocrawl. All commands follow a consistent pattern for dependency injection and error handling.

## Dependency Injection

We use **explicit dependency passing** via the `CommandDeps` struct:

```go
type CommandDeps struct {
    Logger logger.Interface
    Config config.Interface
}
```

### Creating Dependencies

Use the factory function:

```go
deps, err := common.NewCommandDeps()
if err != nil {
    return fmt.Errorf("failed to get dependencies: %w", err)
}
```

### Storage Client

For commands that need storage, use the shared function:

```go
storageClient, err := common.CreateStorageClient(deps.Config, deps.Logger)
```

## Command Structure

Each command follows this pattern:

```go
func Command(deps CommandDeps) *cobra.Command {
    return &cobra.Command{
        Use:   "mycommand",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Get dependencies
            deps, err := common.NewCommandDeps()
            if err != nil {
                return fmt.Errorf("failed to get dependencies: %w", err)
            }
            
            // 2. Create additional dependencies (storage, etc.)
            // ...
            
            // 3. Execute command logic
            return executeCommand(cmd.Context(), deps, args)
        },
    }
}
```

## Error Handling

1. Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
2. Log errors at command level only
3. Return errors for caller to handle

## Constants

Common constants are defined in `cmd/common/constants.go`:

- Index names (DefaultArticleIndex, DefaultPageIndex, etc.)
- Timeouts (DefaultCrawlerTimeout, DefaultShutdownTimeout)
- Capacities (DefaultIndicesCapacity)

## Adding a New Command

1. Create `cmd/mycommand/mycommand.go`
2. Implement command following the pattern above
3. Register in `cmd/root.go` init():

   ```go
   rootCmd.AddCommand(mycommand.Command())
   ```

## Testing

Commands are designed to be testable:

```go
func TestMyCommand(t *testing.T) {
    deps := common.CommandDeps{
        Logger: mockLogger,
        Config: mockConfig,
    }
    
    err := executeCommand(context.Background(), deps, args)
    assert.NoError(t, err)
}
```
