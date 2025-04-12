package logger

import (
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

// LogOnStartExecuting logs the OnStartExecuting event.
func (l *fxLogger) LogOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.log.Debug("OnStart hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

// LogOnStartExecuted logs the OnStartExecuted event.
func (l *fxLogger) LogOnStartExecuted(e *fxevent.OnStartExecuted) {
	if e.Err != nil {
		l.log.Error("OnStart hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("OnStart hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime,
		)
	}
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
	if e.Err != nil {
		l.log.Error("OnStop hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("OnStop hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime,
		)
	}
}

// LogSupplied logs the Supplied event.
func (l *fxLogger) LogSupplied(e *fxevent.Supplied) {
	if e.Err != nil {
		l.log.Error("Error encountered while applying options",
			"type", e.TypeName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("Supplied",
			"type", e.TypeName,
		)
	}
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
	if e.Err != nil {
		l.log.Error("Invoke failed",
			"function", e.FunctionName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("Invoked",
			"function", e.FunctionName,
		)
	}
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
	if e.Err != nil {
		l.log.Error("Custom logger initialization failed",
			"error", e.Err,
		)
	} else {
		l.log.Debug("Initialized custom fxevent.Logger",
			"function", e.ConstructorName,
		)
	}
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
