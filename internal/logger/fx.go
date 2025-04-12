package logger

import (
	"go.uber.org/fx/fxevent"
)

// NewFxLogger creates a new Fx logger that uses our custom logger.
func NewFxLogger(log Interface) fxevent.Logger {
	return &fxLogger{log: log}
}

type fxLogger struct {
	log Interface
}

func (l *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logOnStartExecuting(e)
	case *fxevent.OnStartExecuted:
		l.logOnStartExecuted(e)
	case *fxevent.OnStopExecuting:
		l.logOnStopExecuting(e)
	case *fxevent.OnStopExecuted:
		l.logOnStopExecuted(e)
	case *fxevent.Supplied:
		l.logSupplied(e)
	case *fxevent.Provided:
		l.logProvided(e)
	case *fxevent.Replaced:
		l.logReplaced(e)
	case *fxevent.Decorated:
		l.logDecorated(e)
	case *fxevent.Invoked:
		l.logInvoked(e)
	case *fxevent.Stopped:
		l.logStopped(e)
	case *fxevent.RolledBack:
		l.logRolledBack(e)
	case *fxevent.Started:
		l.logStarted()
	}
}

func (l *fxLogger) logOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.log.Debug("OnStart hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"event", "OnStartExecuting",
	)
}

func (l *fxLogger) logOnStartExecuted(e *fxevent.OnStartExecuted) {
	if e.Err != nil {
		l.log.Error("OnStart hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
			"event", "OnStartExecuted",
		)
	} else {
		l.log.Debug("OnStart hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime.String(),
			"event", "OnStartExecuted",
		)
	}
}

func (l *fxLogger) logOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.log.Debug("OnStop hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
		"event", "OnStopExecuting",
	)
}

func (l *fxLogger) logOnStopExecuted(e *fxevent.OnStopExecuted) {
	if e.Err != nil {
		l.log.Error("OnStop hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
			"event", "OnStopExecuted",
		)
	} else {
		l.log.Debug("OnStop hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime.String(),
			"event", "OnStopExecuted",
		)
	}
}

func (l *fxLogger) logSupplied(e *fxevent.Supplied) {
	if e.Err != nil {
		l.log.Error("Error encountered while applying options",
			"type", e.TypeName,
			"error", e.Err,
			"event", "Supplied",
		)
	} else {
		l.log.Debug("Supplied",
			"type", e.TypeName,
			"event", "Supplied",
		)
	}
}

func (l *fxLogger) logProvided(e *fxevent.Provided) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Provided",
			"constructor", e.ConstructorName,
			"type", rtype,
			"event", "Provided",
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while providing options",
			"constructor", e.ConstructorName,
			"error", e.Err,
			"event", "Provided",
		)
	}
}

func (l *fxLogger) logReplaced(e *fxevent.Replaced) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Replaced",
			"type", rtype,
			"event", "Replaced",
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while replacing options",
			"error", e.Err,
			"event", "Replaced",
		)
	}
}

func (l *fxLogger) logDecorated(e *fxevent.Decorated) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Decorated",
			"decorator", e.DecoratorName,
			"type", rtype,
			"event", "Decorated",
		)
	}
	if e.Err != nil {
		l.log.Error("Error encountered while decorating options",
			"decorator", e.DecoratorName,
			"error", e.Err,
			"event", "Decorated",
		)
	}
}

func (l *fxLogger) logInvoked(e *fxevent.Invoked) {
	if e.Err != nil {
		l.log.Error("Error encountered while invoking function",
			"function", e.FunctionName,
			"error", e.Err,
			"event", "Invoked",
		)
	} else {
		l.log.Debug("Invoked",
			"function", e.FunctionName,
			"event", "Invoked",
		)
	}
}

func (l *fxLogger) logStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.log.Error("Error encountered while stopping",
			"error", e.Err,
			"event", "Stopped",
		)
	} else {
		l.log.Debug("Stopped",
			"event", "Stopped",
		)
	}
}

func (l *fxLogger) logRolledBack(e *fxevent.RolledBack) {
	l.log.Error("Rolled back",
		"error", e.Err,
		"event", "RolledBack",
	)
}

func (l *fxLogger) logStarted() {
	l.log.Debug("Started",
		"event", "Started",
	)
}
