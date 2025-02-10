# Config Component

The Config component manages application configuration using environment variables and command-line flags.

## Interface

```go
type Config struct {
    AppName         string
    AppEnv          string
    AppDebug        bool
    ElasticURL      string
    ElasticPassword string
    ElasticAPIKey   string
    IndexName       string
    LogLevel        string
}
```

## Loading Configuration

```go
func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        // Handle .env loading error
    }
    
    config := &Config{
        AppName:         getEnvOrDefault("APP_NAME", "gocrawl"),
        AppEnv:          getEnvOrDefault("APP_ENV", "development"),
        AppDebug:        getBoolEnvOrDefault("APP_DEBUG", false),
        ElasticURL:      os.Getenv("ELASTIC_URL"),
        ElasticPassword: os.Getenv("ELASTIC_PASSWORD"),
        ElasticAPIKey:   os.Getenv("ELASTIC_API_KEY"),
        IndexName:       os.Getenv("INDEX_NAME"),
        LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
    }
    
    return config, config.Validate()
}
```

## Key Features

### 1. Environment Variables

```go
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 2. Configuration Validation

```go
func (c *Config) Validate() error {
    if c.ElasticURL == "" {
        return errors.New("ELASTIC_URL is required")
    }
    if c.ElasticPassword == "" {
        return errors.New("ELASTIC_PASSWORD is required")
    }
    return nil
}
```

## Best Practices

1. **Security**
   - Never log sensitive values
   - Use secure environment variables
   - Validate all inputs

2. **Defaults**
   - Provide sensible defaults
   - Document all options
   - Use clear naming conventions
