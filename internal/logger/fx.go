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
	l.log.Debug("Fx: OnStart hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

func (l *fxLogger) logOnStartExecuted(e *fxevent.OnStartExecuted) {
	if e.Err != nil {
		l.log.Error("Fx: OnStart hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
		)
	} else {
		l.log.Info("Fx: OnStart hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime.String(),
		)
	}
}

func (l *fxLogger) logOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.log.Debug("Fx: OnStop hook executing",
		"callee", e.FunctionName,
		"caller", e.CallerName,
	)
}

func (l *fxLogger) logOnStopExecuted(e *fxevent.OnStopExecuted) {
	if e.Err != nil {
		l.log.Error("Fx: OnStop hook failed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"error", e.Err,
		)
	} else {
		l.log.Info("Fx: OnStop hook executed",
			"callee", e.FunctionName,
			"caller", e.CallerName,
			"runtime", e.Runtime.String(),
		)
	}
}

func (l *fxLogger) logSupplied(e *fxevent.Supplied) {
	if e.Err != nil {
		l.log.Error("Fx: Error applying options",
			"type", e.TypeName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("Fx: Supplied",
			"type", e.TypeName,
		)
	}
}

func (l *fxLogger) logProvided(e *fxevent.Provided) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Info("Fx: Provided",
			"constructor", e.ConstructorName,
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Fx: Error providing options",
			"constructor", e.ConstructorName,
			"error", e.Err,
		)
	}
}

func (l *fxLogger) logReplaced(e *fxevent.Replaced) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Fx: Replaced",
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Fx: Error replacing options",
			"error", e.Err,
		)
	}
}

func (l *fxLogger) logDecorated(e *fxevent.Decorated) {
	for _, rtype := range e.OutputTypeNames {
		l.log.Debug("Fx: Decorated",
			"decorator", e.DecoratorName,
			"type", rtype,
		)
	}
	if e.Err != nil {
		l.log.Error("Fx: Error decorating options",
			"decorator", e.DecoratorName,
			"error", e.Err,
		)
	}
}

func (l *fxLogger) logInvoked(e *fxevent.Invoked) {
	if e.Err != nil {
		l.log.Error("Fx: Error invoking function",
			"function", e.FunctionName,
			"error", e.Err,
		)
	} else {
		l.log.Debug("Fx: Invoked",
			"function", e.FunctionName,
		)
	}
}

func (l *fxLogger) logStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.log.Error("Fx: Error stopping",
			"error", e.Err,
		)
	} else {
		l.log.Info("Fx: Stopped")
	}
}

func (l *fxLogger) logRolledBack(e *fxevent.RolledBack) {
	l.log.Error("Fx: Rolled back",
		"error", e.Err,
	)
}

func (l *fxLogger) logStarted() {
	l.log.Info("Fx: Started")
}
