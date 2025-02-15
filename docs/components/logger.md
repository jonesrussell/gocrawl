# Logger Component

The Logger component provides structured logging capabilities using Uber's Zap logging framework.

## Interface

```go
type Interface interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Debug(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Fatalf(msg string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

## Configuration

```go
type Params struct {
    Level  zapcore.Level
    AppEnv string
}
```

## Key Features

### 1. Logger Initialization

```go
func NewCustomLogger(params Params) (*CustomLogger, error) {
    var config zap.Config
    
    if params.AppEnv == "development" {
        config = zap.NewDevelopmentConfig()
    } else {
        config = zap.NewProductionConfig()
    }
    
    config.Level = zap.NewAtomicLevelAt(params.Level)
    logger, err := config.Build()
    if err != nil {
        return nil, fmt.Errorf("error initializing logger: %w", err)
    }
    
    return &CustomLogger{Logger: logger}, nil
}
```

### 2. Structured Logging

```go
func (l *CustomLogger) Info(msg string, fields ...interface{}) {
    l.Logger.Info(msg, convertToZapFields(fields...)...)
}

func (l *CustomLogger) Error(msg string, fields ...interface{}) {
    l.Logger.Error(msg, convertToZapFields(fields...)...)
}

func (l *CustomLogger) Debug(msg string, fields ...interface{}) {
    l.Logger.Debug(msg, convertToZapFields(fields...)...)
}

func (l *CustomLogger) Warn(msg string, fields ...interface{}) {
    l.Logger.Warn(msg, convertToZapFields(fields...)...)
}

func (l *CustomLogger) Fatalf(msg string, args ...interface{}) {
    l.Logger.Fatal(fmt.Sprintf(msg, args...))
}

func (l *CustomLogger) Errorf(format string, args ...interface{}) {
    l.Logger.Error(fmt.Sprintf(format, args...))
}
```

## Best Practices

1. **Performance**
   - Use appropriate log levels to avoid excessive logging.
   - Implement sampling in production to reduce log volume.
   - Avoid expensive logging operations in performance-critical paths.

2. **Context**
   - Include relevant context in logs to aid debugging.
   - Use structured logging to maintain consistent field names.
   - Ensure sensitive information is not logged.

3. **Configuration**
   - Provide sensible defaults for logging configurations.
   - Document all logging options clearly.
   - Validate logging configurations at startup.
