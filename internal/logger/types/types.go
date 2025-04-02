package types

// Logger is an interface for structured logging capabilities.
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Printf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Sync() error
}
