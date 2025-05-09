package logger

import (
	"strings"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// fxLogger implements fxevent.Logger
type fxLogger struct {
	log Interface
}

// NewFxLogger creates a new Fx logger
func NewFxLogger(log Interface) fxevent.Logger {
	return &fxLogger{log: log}
}

// isImportantDependency checks if a dependency is important enough to log
func isImportantDependency(name string) bool {
	importantTypes := map[string]bool{
		"*article.Manager":      true,
		"*elasticsearch.Client": true,
		"*sources.Sources":      true,
		"*events.EventBus":      true,
		"signal.Interface":      true,
		"common.JobService":     true,
		"config.Interface":      true,
		"storage.Interface":     true,
		"content.Interface":     true,
		"crawler.Interface":     true,
	}
	return importantTypes[name]
}

// cleanConstructorName cleans up a constructor name for logging
func cleanConstructorName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

// logProvidedEvent logs a Provided event
func (l *fxLogger) logProvidedEvent(event *fxevent.Provided) {
	// Filter out internal dependencies
	if !isImportantDependency(event.ConstructorName) {
		return
	}

	// Clean up constructor name for logging
	constructorName := cleanConstructorName(event.ConstructorName)

	// Log the event
	l.log.Info("provided",
		zap.String("constructor", constructorName),
		zap.Strings("outputs", event.OutputTypeNames),
		zap.String("module", event.ModuleName),
	)
}

// logInvokedEvent logs an Invoked event
func (l *fxLogger) logInvokedEvent(event *fxevent.Invoked) {
	// Clean up function name for logging
	functionName := cleanConstructorName(event.FunctionName)

	// Log the event
	l.log.Info("invoked",
		zap.String("function", functionName),
		zap.String("module", event.ModuleName),
	)
}

// logRunEvent logs a Run event
func (l *fxLogger) logRunEvent(event *fxevent.Run) {
	// Clean up function name for logging
	functionName := cleanConstructorName(event.Name)

	// Log the event
	l.log.Info("run",
		zap.String("function", functionName),
		zap.String("module", event.ModuleName),
	)
}

// logStartEvent logs a Start event
func (l *fxLogger) logStartEvent() {
	l.log.Info("started")
}

// logStopEvent logs a Stop event
func (l *fxLogger) logStopEvent() {
	l.log.Info("stopped")
}

// logRollbackEvent logs a Rollback event
func (l *fxLogger) logRollbackEvent(event *fxevent.RolledBack) {
	l.log.Error("rolled back",
		zap.Error(event.Err),
	)
}

// LogEvent logs an Fx event
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.Provided:
		l.logProvidedEvent(e)
	case *fxevent.Invoked:
		l.logInvokedEvent(e)
	case *fxevent.Run:
		l.logRunEvent(e)
	case *fxevent.Started:
		l.logStartEvent()
	case *fxevent.Stopped:
		l.logStopEvent()
	case *fxevent.RolledBack:
		l.logRollbackEvent(e)
	}
}
