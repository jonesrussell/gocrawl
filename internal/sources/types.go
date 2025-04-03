package sources

// Type aliases for source-related interfaces and types.
type (
	// Logger is an alias for the logger interface, providing
	// structured logging capabilities across the application.
	Logger interface {
		Debug(msg string, fields ...any)
		Info(msg string, fields ...any)
		Warn(msg string, fields ...any)
		Error(msg string, fields ...any)
		Fatal(msg string, fields ...any)
		With(fields ...any) interface {
			Debug(msg string, fields ...any)
			Info(msg string, fields ...any)
			Warn(msg string, fields ...any)
			Error(msg string, fields ...any)
			Fatal(msg string, fields ...any)
			With(fields ...any) interface{}
		}
	}
)
