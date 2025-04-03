package config

// Logger defines the minimal logging interface needed by the config package.
type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

// Field represents a structured logging field.
type Field struct {
	Key   string
	Value any
}

// String creates a Field with a string value.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Error creates a Field with an error value.
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}
