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
    // Error handling...
    
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
```

## Best Practices

1. **Performance**
   - Use appropriate log levels
   - Implement sampling in production
   - Avoid expensive logging operations

2. **Context**
   - Include relevant context in logs
   - Use structured logging
   - Maintain consistent field names
