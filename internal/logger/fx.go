package logger

import (
	"time"

	"go.uber.org/fx/fxevent"
)

// fxLogger is a logger that implements fxevent.Logger.
type fxLogger struct {
	log Interface
}

// NewFxLogger creates a new fx logger.
func NewFxLogger(log Interface) fxevent.Logger {
	return &fxLogger{log: log}
}

// logSuccessOrError is a helper method for logging success or error events.
func (l *fxLogger) logSuccessOrError(successMsg, errorMsg string, err error, fields ...interface{}) {
	if err != nil {
		l.log.Error(errorMsg, append(fields, "error", err)...)
	} else {
		l.log.Debug(successMsg, fields...)
	}
}

// hookEvent represents an event with function and caller information.
type hookEvent interface {
	GetFunctionName() string
	GetCallerName() string
}

// errorEvent represents an event with an error field.
type errorEvent interface {
	GetErr() error
}

// multiTypeEvent represents an event with multiple types and an error.
type multiTypeEvent interface {
	errorEvent
	GetOutputTypeNames() []string
}

// logHookEvent is a helper method for logging hook events (OnStart/OnStop).
func (l *fxLogger) logHookEvent(prefix string, e hookEvent, err error, runtime time.Duration) {
	l.logSuccessOrError(
		prefix+" hook executed",
		prefix+" hook failed",
		err,
		"callee", e.GetFunctionName(),
		"caller", e.GetCallerName(),
		"runtime", runtime,
	)
}

// logTypeEvent is a helper method for logging events with type information.
func (l *fxLogger) logTypeEvent(msg string, e errorEvent, typeName string) {
	l.logSuccessOrError(
		msg,
		"Error encountered while "+msg,
		e.GetErr(),
		"type", typeName,
	)
}

// logMultiTypeEvent is a helper method for logging events with multiple types.
func (l *fxLogger) logMultiTypeEvent(msg string, e multiTypeEvent, additionalFields ...interface{}) {
	for _, rtype := range e.GetOutputTypeNames() {
		l.log.Debug(msg, append(additionalFields, "type", rtype)...)
	}
	if e.GetErr() != nil {
		l.log.Error("Error encountered while "+msg, append(additionalFields, "error", e.GetErr())...)
	}
}

// LogOnStartExecuting logs the OnStartExecuting event.
func (l *fxLogger) LogOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.log.Debug("OnStart hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

// LogOnStartExecuted logs the OnStartExecuted event.
func (l *fxLogger) LogOnStartExecuted(e *fxevent.OnStartExecuted) {
	l.logSuccessOrError(
		"OnStart hook executed",
		"OnStart hook failed",
		e.Err,
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"runtime", e.Runtime,
	)
}

// LogOnStopExecuting logs the OnStopExecuting event.
func (l *fxLogger) LogOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.log.Debug("OnStop hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

// LogOnStopExecuted logs the OnStopExecuted event.
func (l *fxLogger) LogOnStopExecuted(e *fxevent.OnStopExecuted) {
	l.logSuccessOrError(
		"OnStop hook executed",
		"OnStop hook failed",
		e.Err,
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"runtime", e.Runtime,
	)
}

// LogSupplied logs the Supplied event.
func (l *fxLogger) LogSupplied(e *fxevent.Supplied) {
	l.logSuccessOrError(
		"Supplied",
		"Error encountered while applying options",
		e.Err,
		"type", e.TypeName,
	)
}

// LogProvided logs the Provided event.
func (l *fxLogger) LogProvided(e *fxevent.Provided) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Provided",
			"constructor", e.ConstructorName,
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while applying options",
			"error", e.Err,
		)
	}
}

// LogReplaced logs the Replaced event.
func (l *fxLogger) LogReplaced(e *fxevent.Replaced) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Replaced",
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while replacing",
			"error", e.Err,
		)
	}
}

// LogDecorated logs the Decorated event.
func (l *fxLogger) LogDecorated(e *fxevent.Decorated) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Decorated",
			"decorator", e.DecoratorName,
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while applying options",
			"error", e.Err,
		)
	}
}

// LogInvoked logs the Invoked event.
func (l *fxLogger) LogInvoked(e *fxevent.Invoked) {
	l.logSuccessOrError(
		"Invoked",
		"Invoke failed",
		e.Err,
		"function", e.FunctionName,
	)
}

// LogStopping logs the Stopping event.
func (l *fxLogger) LogStopping(e *fxevent.Stopping) {
	l.log.Info("Received signal",
		"signal", e.Signal,
	)
}

// LogStopped logs the Stopped event.
func (l *fxLogger) LogStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.log.Error("Stop failed",
			"error", e.Err,
		)
	}
}

// LogRolledBack logs the RolledBack event.
func (l *fxLogger) LogRolledBack(e *fxevent.RolledBack) {
	l.log.Error("Start failed, rolling back",
		"error", e.Err,
	)
}

// LogStarted logs the Started event.
func (l *fxLogger) LogStarted(e *fxevent.Started) {
	if e.Err != nil {
		l.log.Error("Start failed",
			"error", e.Err,
		)
	} else {
		l.log.Info("Started")
	}
}

// LogLoggerInitialized logs the LoggerInitialized event.
func (l *fxLogger) LogLoggerInitialized(e *fxevent.LoggerInitialized) {
	l.logSuccessOrError(
		"Initialized custom fxevent.Logger",
		"Custom logger initialization failed",
		e.Err,
		"function", e.ConstructorName,
	)
}

// LogEvent logs an fx event.
func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.LogOnStartExecuting(e)
	case *fxevent.OnStartExecuted:
		l.LogOnStartExecuted(e)
	case *fxevent.OnStopExecuting:
		l.LogOnStopExecuting(e)
	case *fxevent.OnStopExecuted:
		l.LogOnStopExecuted(e)
	case *fxevent.Supplied:
		l.LogSupplied(e)
	case *fxevent.Provided:
		l.LogProvided(e)
	case *fxevent.Replaced:
		l.LogReplaced(e)
	case *fxevent.Decorated:
		l.LogDecorated(e)
	case *fxevent.Invoked:
		l.LogInvoked(e)
	case *fxevent.Stopping:
		l.LogStopping(e)
	case *fxevent.Stopped:
		l.LogStopped(e)
	case *fxevent.RolledBack:
		l.LogRolledBack(e)
	case *fxevent.Started:
		l.LogStarted(e)
	case *fxevent.LoggerInitialized:
		l.LogLoggerInitialized(e)
	}
}
